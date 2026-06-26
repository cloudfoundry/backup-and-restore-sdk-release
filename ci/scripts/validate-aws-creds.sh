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
#   ./ci/scripts/validate-aws-creds.sh
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

ok()   { echo "  PASS  $*"; PASS=$((PASS+1)); }
fail() { echo "  FAIL  $*"; FAIL=$((FAIL+1)); }
warn() { echo "  WARN  $*"; WARN=$((WARN+1)); }
header() { echo ""; echo "=== $* ==="; }

check() {
  # check "description" cmd [args...]
  local desc="$1"; shift
  if "$@" 2>&1 >/dev/null; then
    ok "${desc}"
  else
    fail "${desc}"
  fi
}

# ---------------------------------------------------------------------------
# Optional: assume role
# ---------------------------------------------------------------------------
if [[ -n "${AWS_ROLE_ARN:-}" ]]; then
  echo "Assuming role: ${AWS_ROLE_ARN}"
  creds_json=$(aws sts assume-role \
    --role-arn "${AWS_ROLE_ARN}" \
    --role-session-name sdk-validation \
    --query 'Credentials' \
    --output json) || { echo "  FAIL  sts:AssumeRole"; exit 1; }
  export AWS_ACCESS_KEY_ID=$(echo "${creds_json}"    | jq -r '.AccessKeyId')
  export AWS_SECRET_ACCESS_KEY=$(echo "${creds_json}" | jq -r '.SecretAccessKey')
  export AWS_SESSION_TOKEN=$(echo "${creds_json}"    | jq -r '.SessionToken')
  echo "  Role assumed successfully"
fi

# ---------------------------------------------------------------------------
# 1. Identity
# ---------------------------------------------------------------------------
header "1. Caller identity"
identity=$(aws sts get-caller-identity --output json)
account=$(echo "${identity}" | jq -r '.Account')
arn=$(echo "${identity}" | jq -r '.Arn')
echo "  Account : ${account}"
echo "  ARN     : ${arn}"
ok "sts:GetCallerIdentity"

# ---------------------------------------------------------------------------
# 2. S3 pipeline state bucket
# ---------------------------------------------------------------------------
header "2. S3 pipeline state bucket (${S3_PIPELINE_BUCKET})"

check "s3:HeadBucket ${S3_PIPELINE_BUCKET}" \
  aws s3api head-bucket --bucket "${S3_PIPELINE_BUCKET}" --region "${REGION}"

TMPKEY=".sdk-cred-validation-$$"
if aws s3 cp /dev/null "s3://${S3_PIPELINE_BUCKET}/${TMPKEY}" \
     --region "${REGION}" 2>&1 >/dev/null; then
  aws s3 rm "s3://${S3_PIPELINE_BUCKET}/${TMPKEY}" \
    --region "${REGION}" 2>&1 >/dev/null || true
  ok "s3:PutObject + s3:DeleteObject on ${S3_PIPELINE_BUCKET}"
else
  fail "s3:PutObject on ${S3_PIPELINE_BUCKET} — write access missing"
fi

check "s3:GetBucketVersioning ${S3_PIPELINE_BUCKET}" \
  aws s3api get-bucket-versioning --bucket "${S3_PIPELINE_BUCKET}" --region "${REGION}"

# ---------------------------------------------------------------------------
# 3. S3 blobstore test buckets
# ---------------------------------------------------------------------------
header "3. S3 blobstore test buckets"
for entry in "${S3_BLOBSTORE_BUCKETS[@]}"; do
  bucket="${entry%%:*}"
  region="${entry##*:}"
  check "s3:HeadBucket ${bucket} (${region})" \
    aws s3api head-bucket --bucket "${bucket}" --region "${region}"
done

# ---------------------------------------------------------------------------
# 4. S3: terraform state key
# ---------------------------------------------------------------------------
header "4. S3 terraform state (${S3_PIPELINE_BUCKET}/${TERRAFORM_STATE_KEY})"
if aws s3api get-object \
     --bucket "${S3_PIPELINE_BUCKET}" \
     --key "${TERRAFORM_STATE_KEY}" \
     --region "${REGION}" \
     /tmp/sdk-tf-state-$$ 2>&1 >/dev/null; then
  rm -f /tmp/sdk-tf-state-$$
  ok "Existing terraform state found"
else
  warn "No existing terraform state (expected on first run — not an error)"
fi

# ---------------------------------------------------------------------------
# 5. EC2 — bbl-on-AWS needs ec2 + vpc + iam + elb
# ---------------------------------------------------------------------------
header "5. EC2 (required for bbl-on-AWS / system-tests-s3-iam-instance-profile)"

check "ec2:DescribeAvailabilityZones (${REGION})" \
  aws ec2 describe-availability-zones --region "${REGION}"

check "ec2:DescribeVpcs" \
  aws ec2 describe-vpcs --region "${REGION}"

check "ec2:DescribeSubnets" \
  aws ec2 describe-subnets --region "${REGION}"

check "ec2:DescribeSecurityGroups" \
  aws ec2 describe-security-groups --region "${REGION}"

check "ec2:DescribeKeyPairs" \
  aws ec2 describe-key-pairs --region "${REGION}"

check "ec2:DescribeInstances" \
  aws ec2 describe-instances --region "${REGION}"

# ---------------------------------------------------------------------------
# 6. IAM — bbl-on-AWS needs IAM for instance profiles
# ---------------------------------------------------------------------------
header "6. IAM (required for bbl-on-AWS + instance profile attachment)"

check "iam:ListInstanceProfiles" \
  aws iam list-instance-profiles

check "iam:ListRoles" \
  aws iam list-roles

# ---------------------------------------------------------------------------
# 7. RDS — terraform for system-tests-external-dbs-rds
# ---------------------------------------------------------------------------
header "7. RDS (required for system-tests-external-dbs-rds)"

check "rds:DescribeDBInstances (${REGION})" \
  aws rds describe-db-instances --region "${REGION}"

check "rds:DescribeDBEngineVersions (postgres)" \
  aws rds describe-db-engine-versions --engine postgres --engine-version 16 --region "${REGION}"

check "rds:DescribeDBSubnetGroups" \
  aws rds describe-db-subnet-groups --region "${REGION}"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "========================================="
echo "  Results: ${PASS} passed, ${FAIL} failed, ${WARN} warnings"
echo "========================================="

if (( FAIL > 0 )); then
  echo ""
  echo "  At least one permission check failed."
  echo "  OPTIONS:"
  echo "    A) Grant the missing permissions to the bbr_cli_aws_creds role/user"
  echo "       and reuse that credential in the SDK pipeline (bbr_cli_aws_creds.*)"
  echo "    B) Create a separate 'bbr_aws_access_key' credential in CFF CredHub"
  echo "       with the required permissions"
  exit 1
else
  echo ""
  echo "  All required permissions confirmed for account ${account}."
  echo "  You can reuse bbr_cli_aws_creds — update the SDK pipeline to use:"
  echo "    ((bbr_cli_aws_creds.access_key_id))   instead of ((bbr_aws_access_key.username))"
  echo "    ((bbr_cli_aws_creds.secret_access_key)) instead of ((bbr_aws_access_key.password))"
  echo "    ((bbr_cli_aws_creds.role_arn))          (same name)"
  exit 0
fi
