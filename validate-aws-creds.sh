#!/usr/bin/env bash
# validate-aws-creds.sh
# ---------------------------------------------------------------------------
# Run this script with the credentials you want to validate (bbr_cli_aws_creds
# or any other set).  It checks every AWS permission the SDK pipeline needs,
# reports PASS/FAIL for each requirement, and exits non-zero if any fail.
#
# Usage:
#   AWS_ACCESS_KEY_ID=xxx \
#   AWS_SECRET_ACCESS_KEY=yyy \
#   [AWS_ROLE_ARN=zzz] \        # optional: assume-role ARN
#   ./validate-aws-creds.sh
# ---------------------------------------------------------------------------
set -euo pipefail

REGION="eu-west-1"
REGION_BACKUP="eu-central-1"
S3_PIPELINE_BUCKET="backup-and-restore-sdk-releases"
S3_BLOBSTORE_BUCKETS=(
  "large-blob-test-bucket-unversioned:${REGION}"
  "bbr-system-test-bucket:${REGION}"
  "bbr-system-test-bucket-clone:${REGION_BACKUP}"
  "bbr-system-test-s3-unversioned-bucket:${REGION}"
)
TERRAFORM_STATE_KEY="terraform-rds-state"

PASS=0
FAIL=0
WARN=0

ok()   { echo "  PASS  $*"; ((PASS++)); }
fail() { echo "  FAIL  $*"; ((FAIL++)); }
warn() { echo "  WARN  $*"; ((WARN++)); }
header() { echo ""; echo "=== $* ==="; }

# ---------------------------------------------------------------------------
# Optional: assume role
# ---------------------------------------------------------------------------
if [[ -n "${AWS_ROLE_ARN:-}" ]]; then
  echo "Assuming role: ${AWS_ROLE_ARN}"
  creds_json=$(aws sts assume-role \
    --role-arn "${AWS_ROLE_ARN}" \
    --role-session-name sdk-validation \
    --query 'Credentials' \
    --output json 2>&1) || { echo "  FAIL  sts:AssumeRole (${creds_json})"; exit 1; }
  export AWS_ACCESS_KEY_ID=$(echo "${creds_json}" | jq -r '.AccessKeyId')
  export AWS_SECRET_ACCESS_KEY=$(echo "${creds_json}" | jq -r '.SecretAccessKey')
  export AWS_SESSION_TOKEN=$(echo "${creds_json}" | jq -r '.SessionToken')
  echo "  Role assumed successfully"
fi

# ---------------------------------------------------------------------------
# 1. Identity
# ---------------------------------------------------------------------------
header "1. Caller identity"
identity=$(aws sts get-caller-identity --output json 2>&1) \
  && { account=$(echo "${identity}" | jq -r '.Account'); arn=$(echo "${identity}" | jq -r '.Arn')
       echo "  Account : ${account}"
       echo "  ARN     : ${arn}"
       ok "sts:GetCallerIdentity"; } \
  || fail "sts:GetCallerIdentity (${identity})"

# ---------------------------------------------------------------------------
# 2. S3 pipeline state bucket
# ---------------------------------------------------------------------------
header "2. S3 pipeline state bucket (${S3_PIPELINE_BUCKET})"

aws s3api head-bucket --bucket "${S3_PIPELINE_BUCKET}" --region "${REGION}" 2>&1 \
  && ok "s3:HeadBucket ${S3_PIPELINE_BUCKET}" \
  || fail "s3:HeadBucket ${S3_PIPELINE_BUCKET} — bucket may not exist or no access"

TMPFILE="sdk-cred-validation-$$"
aws s3 cp /dev/null "s3://${S3_PIPELINE_BUCKET}/.${TMPFILE}" \
  --region "${REGION}" 2>&1 >/dev/null \
  && { aws s3 rm "s3://${S3_PIPELINE_BUCKET}/.${TMPFILE}" \
         --region "${REGION}" 2>&1 >/dev/null
       ok "s3:PutObject + s3:DeleteObject on ${S3_PIPELINE_BUCKET}"; } \
  || fail "s3:PutObject on ${S3_PIPELINE_BUCKET} — write access missing"

aws s3api get-bucket-versioning --bucket "${S3_PIPELINE_BUCKET}" \
  --region "${REGION}" 2>&1 >/dev/null \
  && ok "s3:GetBucketVersioning ${S3_PIPELINE_BUCKET}" \
  || warn "s3:GetBucketVersioning ${S3_PIPELINE_BUCKET} — needed for versioned dev-release resource"

# ---------------------------------------------------------------------------
# 3. S3 blobstore test buckets
# ---------------------------------------------------------------------------
header "3. S3 blobstore test buckets"
for entry in "${S3_BLOBSTORE_BUCKETS[@]}"; do
  bucket="${entry%%:*}"
  region="${entry##*:}"
  aws s3api head-bucket --bucket "${bucket}" --region "${region}" 2>&1 >/dev/null \
    && ok "s3:HeadBucket ${bucket} (${region})" \
    || fail "s3:HeadBucket ${bucket} — bucket missing or no access"
done

# ---------------------------------------------------------------------------
# 4. S3: terraform state key
# ---------------------------------------------------------------------------
header "4. S3 terraform state (${S3_PIPELINE_BUCKET}/${TERRAFORM_STATE_KEY})"
aws s3api get-object \
  --bucket "${S3_PIPELINE_BUCKET}" \
  --key "${TERRAFORM_STATE_KEY}" \
  --region "${REGION}" \
  /tmp/sdk-tf-state-$$ 2>&1 >/dev/null \
  && { rm -f /tmp/sdk-tf-state-$$; ok "Existing terraform state found"; } \
  || warn "No existing terraform state (expected on first run — not an error)"

# ---------------------------------------------------------------------------
# 5. EC2 — bbl-on-AWS needs ec2 + vpc + iam + elb + route53
# ---------------------------------------------------------------------------
header "5. EC2 (required for bbl-on-AWS, system-tests-s3-iam-instance-profile)"

aws ec2 describe-availability-zones --region "${REGION}" \
  --query 'AvailabilityZones[0].ZoneName' --output text 2>&1 \
  && ok "ec2:DescribeAvailabilityZones" \
  || fail "ec2:DescribeAvailabilityZones"

aws ec2 describe-vpcs --region "${REGION}" \
  --query 'length(Vpcs)' --output text 2>&1 >/dev/null \
  && ok "ec2:DescribeVpcs" \
  || fail "ec2:DescribeVpcs"

aws ec2 describe-subnets --region "${REGION}" \
  --query 'length(Subnets)' --output text 2>&1 >/dev/null \
  && ok "ec2:DescribeSubnets" \
  || fail "ec2:DescribeSubnets"

aws ec2 describe-security-groups --region "${REGION}" \
  --query 'length(SecurityGroups)' --output text 2>&1 >/dev/null \
  && ok "ec2:DescribeSecurityGroups" \
  || fail "ec2:DescribeSecurityGroups"

aws ec2 describe-key-pairs --region "${REGION}" 2>&1 >/dev/null \
  && ok "ec2:DescribeKeyPairs" \
  || fail "ec2:DescribeKeyPairs"

aws ec2 describe-instances --region "${REGION}" \
  --query 'length(Reservations)' --output text 2>&1 >/dev/null \
  && ok "ec2:DescribeInstances" \
  || fail "ec2:DescribeInstances"

# ---------------------------------------------------------------------------
# 6. IAM — bbl-on-AWS needs IAM for roles, instance profiles
# ---------------------------------------------------------------------------
header "6. IAM (required for bbl-on-AWS + instance profile attachment)"

aws iam list-instance-profiles \
  --query 'length(InstanceProfiles)' --output text 2>&1 >/dev/null \
  && ok "iam:ListInstanceProfiles" \
  || fail "iam:ListInstanceProfiles"

aws iam list-roles \
  --query 'length(Roles)' --output text 2>&1 >/dev/null \
  && ok "iam:ListRoles" \
  || fail "iam:ListRoles"

# ---------------------------------------------------------------------------
# 7. RDS — terraform for system-tests-external-dbs-rds
# ---------------------------------------------------------------------------
header "7. RDS (required for system-tests-external-dbs-rds)"

aws rds describe-db-instances --region "${REGION}" \
  --query 'length(DBInstances)' --output text 2>&1 >/dev/null \
  && ok "rds:DescribeDBInstances (${REGION})" \
  || fail "rds:DescribeDBInstances — no RDS permissions"

aws rds describe-db-engine-versions \
  --engine postgres --engine-version 16 \
  --region "${REGION}" \
  --query 'length(DBEngineVersions)' --output text 2>&1 >/dev/null \
  && ok "rds:DescribeDBEngineVersions postgres-16" \
  || fail "rds:DescribeDBEngineVersions"

aws rds describe-db-subnet-groups --region "${REGION}" \
  --query 'length(DBSubnetGroups)' --output text 2>&1 >/dev/null \
  && ok "rds:DescribeDBSubnetGroups" \
  || warn "rds:DescribeDBSubnetGroups — might be needed for VPC-based RDS"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "========================================="
echo "  Results: ${PASS} passed, ${FAIL} failed, ${WARN} warnings"
echo "========================================="

if (( FAIL > 0 )); then
  echo "  ACTION REQUIRED: Missing permissions above must be granted"
  echo "  OR create a separate bbr_aws_access_key credential in CFF CredHub"
  exit 1
else
  echo "  All required permissions confirmed."
  echo "  You can reuse bbr_cli_aws_creds by updating the SDK pipeline."
  exit 0
fi
