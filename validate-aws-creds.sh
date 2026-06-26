#!/usr/bin/env bash
# validate-aws-creds.sh
# ---------------------------------------------------------------------------
# Validates that an AWS credential set has all permissions and all S3/RDS/IAM
# resources the backup-and-restore-sdk-release pipeline requires.
#
# Usage:
#   AWS_ACCESS_KEY_ID=xxx \
#   AWS_SECRET_ACCESS_KEY=yyy \
#   [AWS_ROLE_ARN=zzz] \        # optional: assume-role ARN (set if using bbr_cli_aws_creds)
#   ./validate-aws-creds.sh
#
#   Pass --create-missing to print AWS CLI commands that create any
#   missing resources (buckets, RDS subnet groups, IAM profile scaffolding).
# ---------------------------------------------------------------------------
set -uo pipefail
# NOTE: intentionally NOT set -e — we want all checks to run before reporting.

CREATE_MISSING=false
[[ "${1:-}" == "--create-missing" ]] && CREATE_MISSING=true

REGION="eu-west-1"
REGION_BACKUP="eu-central-1"
REGION_US_EAST="us-east-1"

# ---------------------------------------------------------------------------
# Counter helpers — MUST use PASS=$((PASS+1)) not ((PASS++)).
# ((PASS++)) returns exit code 1 when PASS=0 (arithmetic 0 == false in bash),
# which breaks the surrounding || / && chains even without set -e.
# ---------------------------------------------------------------------------
PASS=0
FAIL=0
WARN=0
MISSING_BUCKETS=()
MISSING_VERSIONED_BUCKETS=()
MISSING_RDS_INSTANCES=()

ok()   { echo "  PASS  $*"; PASS=$((PASS+1)); }
fail() { echo "  FAIL  $*"; FAIL=$((FAIL+1)); }
warn() { echo "  WARN  $*"; WARN=$((WARN+1)); }
header() { echo ""; echo "=== $* ==="; }

# ---------------------------------------------------------------------------
# Optional: assume role (needed when using bbr_cli_aws_creds)
# ---------------------------------------------------------------------------
if [[ -n "${AWS_ROLE_ARN:-}" ]]; then
  echo "Assuming role: ${AWS_ROLE_ARN}"
  creds_json=$(aws sts assume-role \
    --role-arn "${AWS_ROLE_ARN}" \
    --role-session-name sdk-validation \
    --query 'Credentials' \
    --output json 2>&1) || { echo "  FAIL  sts:AssumeRole (${creds_json})"; exit 1; }
  export AWS_ACCESS_KEY_ID
  AWS_ACCESS_KEY_ID=$(echo "${creds_json}" | jq -r '.AccessKeyId')
  export AWS_SECRET_ACCESS_KEY
  AWS_SECRET_ACCESS_KEY=$(echo "${creds_json}" | jq -r '.SecretAccessKey')
  export AWS_SESSION_TOKEN
  AWS_SESSION_TOKEN=$(echo "${creds_json}" | jq -r '.SessionToken')
  echo "  Role assumed successfully"
fi

# ---------------------------------------------------------------------------
# 1. Identity
# ---------------------------------------------------------------------------
header "1. Caller identity"
identity=$(aws sts get-caller-identity --output json 2>&1)
if [[ $? -eq 0 ]]; then
  account=$(echo "${identity}" | jq -r '.Account')
  arn=$(echo "${identity}" | jq -r '.Arn')
  echo "  Account : ${account}"
  echo "  ARN     : ${arn}"
  ok "sts:GetCallerIdentity"
else
  fail "sts:GetCallerIdentity (${identity})"
fi

# ---------------------------------------------------------------------------
# Helper: check a single S3 bucket (exists + optionally versioned)
# ---------------------------------------------------------------------------
check_bucket() {
  local bucket="$1"
  local region="$2"
  local versioned="${3:-false}"   # "true" if the pipeline needs versioning enabled

  if aws s3api head-bucket --bucket "${bucket}" --region "${region}" 2>/dev/null; then
    ok "s3: ${bucket} (${region}) — exists"
    if [[ "${versioned}" == "true" ]]; then
      status=$(aws s3api get-bucket-versioning \
        --bucket "${bucket}" --region "${region}" \
        --query 'Status' --output text 2>/dev/null)
      if [[ "${status}" == "Enabled" ]]; then
        ok "s3: ${bucket} — versioning enabled"
      else
        fail "s3: ${bucket} — versioning NOT enabled (status: ${status:-none})"
        MISSING_VERSIONED_BUCKETS+=("${bucket}:${region}")
      fi
    fi
  else
    fail "s3: ${bucket} (${region}) — bucket MISSING or no access"
    if [[ "${versioned}" == "true" ]]; then
      MISSING_BUCKETS+=("VERSIONED:${bucket}:${region}")
    else
      MISSING_BUCKETS+=("UNVERSIONED:${bucket}:${region}")
    fi
  fi
}

# ---------------------------------------------------------------------------
# 2. Pipeline state + release buckets (VERSIONED — Concourse semver + s3 resources)
# ---------------------------------------------------------------------------
header "2. Pipeline infrastructure buckets (VERSIONED)"
check_bucket "backup-and-restore-sdk-releases"    "${REGION}"  "true"
check_bucket "backup-and-restore-sdk-release-blobs" "${REGION}" "true"

# verify write access on the primary bucket
TMPOBJ="sdk-cred-validation-$$"
if aws s3 cp /dev/null "s3://backup-and-restore-sdk-releases/.${TMPOBJ}" \
     --region "${REGION}" 2>/dev/null; then
  aws s3 rm "s3://backup-and-restore-sdk-releases/.${TMPOBJ}" \
    --region "${REGION}" 2>/dev/null || true
  ok "s3: backup-and-restore-sdk-releases — write access confirmed"
else
  fail "s3: backup-and-restore-sdk-releases — write access DENIED (PutObject)"
fi

# ---------------------------------------------------------------------------
# 3. S3 contract test bucket (UNVERSIONED)
# ---------------------------------------------------------------------------
header "3. S3 contract test bucket (UNVERSIONED)"
check_bucket "large-blob-test-bucket-unversioned" "${REGION}" "false"

# ---------------------------------------------------------------------------
# 4. S3 blobstore system test buckets (system-tests-blobstore)
# ---------------------------------------------------------------------------
header "4. S3 blobstore system test buckets"

# Versioned buckets (used by s3-versioned-blobstore-backup-restorer)
check_bucket "bbr-system-test-bucket"             "${REGION}"        "true"
check_bucket "bbr-system-test-bucket-clone"       "${REGION_BACKUP}" "true"

# Unversioned buckets
check_bucket "bbr-system-test-bucket-unversioned"          "${REGION}"        "false"
check_bucket "bbr-system-test-s3-unversioned-bucket"       "${REGION}"        "false"
check_bucket "bbr-system-test-s3-unversioned-backup-bucket" "${REGION_US_EAST}" "false"
check_bucket "sdk-system-test-unversioned-bpm"             "${REGION}"        "false"
check_bucket "sdk-system-test-unversioned-bpm-backup"      "${REGION}"        "false"
check_bucket "sdk-large-number-of-files"                   "${REGION}"        "false"
check_bucket "sdk-large-number-of-files-backup"            "${REGION}"        "false"
check_bucket "sdk-unversioned-clone"                       "${REGION_US_EAST}" "false"

# ---------------------------------------------------------------------------
# 5. IAM instance profile test buckets (system-tests-s3-iam-instance-profile)
#    These must be VERSIONED (uses s3-versioned-blobstore-backup-restorer via IAM profile)
# ---------------------------------------------------------------------------
header "5. IAM instance profile test buckets (VERSIONED)"
check_bucket "iam-instance-role-test"        "${REGION}" "true"
check_bucket "iam-instance-role-test-clone"  "${REGION}" "true"

# ---------------------------------------------------------------------------
# 6. EC2 permissions (required for bbl up --iaas aws)
# ---------------------------------------------------------------------------
header "6. EC2 permissions (bbl-on-AWS for system-tests-s3-iam-instance-profile)"

ec2_check() {
  local desc="$1"; shift
  if "$@" 2>/dev/null 1>/dev/null; then
    ok "ec2: ${desc}"
  else
    fail "ec2: ${desc} — permission denied"
  fi
}

ec2_check "DescribeAvailabilityZones" \
  aws ec2 describe-availability-zones --region "${REGION}" \
    --query 'AvailabilityZones[0].ZoneName' --output text
ec2_check "DescribeVpcs" \
  aws ec2 describe-vpcs --region "${REGION}" --query 'length(Vpcs)' --output text
ec2_check "DescribeSubnets" \
  aws ec2 describe-subnets --region "${REGION}" --query 'length(Subnets)' --output text
ec2_check "DescribeSecurityGroups" \
  aws ec2 describe-security-groups --region "${REGION}" \
    --query 'length(SecurityGroups)' --output text
ec2_check "DescribeKeyPairs" \
  aws ec2 describe-key-pairs --region "${REGION}"
ec2_check "DescribeInstances" \
  aws ec2 describe-instances --region "${REGION}" \
    --query 'length(Reservations)' --output text
ec2_check "DescribeAddresses (elastic IPs)" \
  aws ec2 describe-addresses --region "${REGION}" --query 'length(Addresses)' --output text
ec2_check "DescribeInternetGateways" \
  aws ec2 describe-internet-gateways --region "${REGION}" \
    --query 'length(InternetGateways)' --output text
ec2_check "DescribeRouteTables" \
  aws ec2 describe-route-tables --region "${REGION}" \
    --query 'length(RouteTables)' --output text
ec2_check "DescribeVolumes" \
  aws ec2 describe-volumes --region "${REGION}" --query 'length(Volumes)' --output text
ec2_check "DescribeImages (AMI listing)" \
  aws ec2 describe-images --region "${REGION}" --owners self --query 'length(Images)' --output text

# ELB (bbl creates a classic load balancer for the BOSH director)
if aws elbv2 describe-load-balancers --region "${REGION}" \
     --query 'length(LoadBalancers)' --output text 2>/dev/null 1>/dev/null; then
  ok "elbv2: DescribeLoadBalancers"
else
  warn "elbv2: DescribeLoadBalancers — may be needed by bbl for BOSH director LB"
fi

# ---------------------------------------------------------------------------
# 7. IAM permissions (bbl-on-AWS creates IAM roles for the BOSH director)
# ---------------------------------------------------------------------------
header "7. IAM permissions (bbl-on-AWS + instance profile for s3-iam tests)"

iam_check() {
  local desc="$1"; shift
  if "$@" 2>/dev/null 1>/dev/null; then
    ok "iam: ${desc}"
  else
    fail "iam: ${desc} — permission denied"
  fi
}

iam_check "ListRoles" \
  aws iam list-roles --query 'length(Roles)' --output text
iam_check "ListInstanceProfiles" \
  aws iam list-instance-profiles --query 'length(InstanceProfiles)' --output text
iam_check "ListPolicies (own)" \
  aws iam list-policies --scope Local --query 'length(Policies)' --output text

# bbl creates roles; we can detect if the current identity is allowed to simulate
# iam:PassRole by attempting a dry-run (iam:SimulatePrincipalPolicy)
principal_arn="${arn:-}"
if [[ -n "${principal_arn}" ]]; then
  sim_result=$(aws iam simulate-principal-policy \
    --policy-source-arn "${principal_arn}" \
    --action-names 'iam:CreateRole' \
    --query 'EvaluationResults[0].EvalDecision' \
    --output text 2>/dev/null) || sim_result="unknown"
  if [[ "${sim_result}" == "allowed" ]]; then
    ok "iam: CreateRole (simulated — allowed)"
  elif [[ "${sim_result}" == "implicitDeny" || "${sim_result}" == "explicitDeny" ]]; then
    fail "iam: CreateRole (simulated — ${sim_result}); bbl-on-AWS will fail"
  else
    warn "iam: CreateRole simulation inconclusive (${sim_result}) — verify manually"
  fi
fi

# ---------------------------------------------------------------------------
# 8. RDS permissions + instance existence
# ---------------------------------------------------------------------------
header "8. RDS permissions + instance status (system-tests-external-dbs-rds)"

if aws rds describe-db-instances --region "${REGION}" \
     --query 'length(DBInstances)' --output text 2>/dev/null 1>/dev/null; then
  ok "rds: DescribeDBInstances"
else
  fail "rds: DescribeDBInstances — no RDS permissions at all"
fi

if aws rds describe-db-subnet-groups --region "${REGION}" \
     --query 'length(DBSubnetGroups)' --output text 2>/dev/null 1>/dev/null; then
  ok "rds: DescribeDBSubnetGroups"
else
  warn "rds: DescribeDBSubnetGroups — Terraform may need this for VPC-based RDS"
fi

# Check that each required RDS instance exists and is available
check_rds() {
  local id="$1"
  local engine="$2"
  local ver="$3"

  status=$(aws rds describe-db-instances \
    --db-instance-identifier "${id}" \
    --region "${REGION}" \
    --query 'DBInstances[0].DBInstanceStatus' \
    --output text 2>/dev/null) || status="NOT_FOUND"

  if [[ "${status}" == "available" ]]; then
    ok "rds: ${id} (${engine} ${ver}) — available"
  elif [[ "${status}" == "NOT_FOUND" || "${status}" == "None" ]]; then
    fail "rds: ${id} (${engine} ${ver}) — DOES NOT EXIST (Terraform will create it)"
    MISSING_RDS_INSTANCES+=("${id}:${engine}:${ver}")
  else
    warn "rds: ${id} — status: ${status} (may be starting up)"
  fi
}

check_rds "postgres-13-system-tests"  "postgres" "13"
check_rds "postgres-15-system-tests"  "postgres" "15"
check_rds "postgres-16-system-tests"  "postgres" "16"
check_rds "mariadb-10-6-system-tests" "mariadb"  "10.6"

# ---------------------------------------------------------------------------
# 9. Terraform state (S3 backend)
# ---------------------------------------------------------------------------
header "9. Terraform state key (backend: s3://backup-and-restore-sdk-releases/terraform-rds-state)"
if aws s3api get-object \
     --bucket "backup-and-restore-sdk-releases" \
     --key "terraform-rds-state" \
     --region "${REGION}" \
     "/tmp/sdk-tf-state-$$" 2>/dev/null 1>/dev/null; then
  rm -f "/tmp/sdk-tf-state-$$"
  ok "Existing terraform state found at terraform-rds-state"
else
  warn "No terraform state yet (expected on first run — Terraform will create it)"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "========================================="
echo "  Results: ${PASS} passed  |  ${FAIL} failed  |  ${WARN} warnings"
echo "========================================="

# ---------------------------------------------------------------------------
# Resource creation commands for missing items
# ---------------------------------------------------------------------------
if [[ ${#MISSING_BUCKETS[@]} -gt 0 ]] || [[ ${#MISSING_VERSIONED_BUCKETS[@]} -gt 0 ]] || [[ ${#MISSING_RDS_INSTANCES[@]} -gt 0 ]]; then
  echo ""
  echo "--- RESOURCE CREATION COMMANDS ---"
  echo "# Run the following to create missing AWS resources in account ${account:-<account>}."
  echo "# You must have the same credentials exported (with role assumed) before running."
  echo ""

  for entry in "${MISSING_BUCKETS[@]}"; do
    IFS=':' read -r vtype bucket region <<< "${entry}"
    if [[ "${vtype}" == "VERSIONED" ]]; then
      if [[ "${region}" == "us-east-1" ]]; then
        echo "aws s3api create-bucket --bucket '${bucket}' --region '${region}'"
      else
        echo "aws s3api create-bucket --bucket '${bucket}' --region '${region}' \\"
        echo "  --create-bucket-configuration LocationConstraint='${region}'"
      fi
      echo "aws s3api put-bucket-versioning --bucket '${bucket}' --region '${region}' \\"
      echo "  --versioning-configuration Status=Enabled"
      echo ""
    else
      if [[ "${region}" == "us-east-1" ]]; then
        echo "aws s3api create-bucket --bucket '${bucket}' --region '${region}'"
      else
        echo "aws s3api create-bucket --bucket '${bucket}' --region '${region}' \\"
        echo "  --create-bucket-configuration LocationConstraint='${region}'"
      fi
      echo ""
    fi
  done

  for entry in "${MISSING_VERSIONED_BUCKETS[@]}"; do
    IFS=':' read -r bucket region <<< "${entry}"
    echo "# Enable versioning on existing bucket:"
    echo "aws s3api put-bucket-versioning --bucket '${bucket}' --region '${region}' \\"
    echo "  --versioning-configuration Status=Enabled"
    echo ""
  done

  for entry in "${MISSING_RDS_INSTANCES[@]}"; do
    IFS=':' read -r id engine ver <<< "${entry}"
    echo "# NOTE: RDS instances are created by Terraform (ci/terraform/bbr-sdk-system-tests/aws/)."
    echo "# Run 'terraform apply' in that directory — it will create '${id}' (${engine} ${ver})."
    break
  done

  # IAM instance profile scaffold (always show if we're printing creation commands)
  echo "# IAM instance profile for system-tests-s3-iam-instance-profile:"
  echo "# (Replace sdk-s3-iam-backuper-profile with your desired profile name)"
  cat <<'IAMEOF'
PROFILE_NAME="sdk-s3-iam-backuper-profile"
BUCKET_NAME="iam-instance-role-test"
REGION="eu-west-1"

# 1. Create the IAM policy
aws iam create-policy \
  --policy-name "sdk-s3-iam-backuper-policy" \
  --policy-document "{
    \"Version\": \"2012-10-17\",
    \"Statement\": [{
      \"Effect\": \"Allow\",
      \"Action\": [
        \"s3:GetObject\",\"s3:PutObject\",\"s3:DeleteObject\",
        \"s3:ListBucket\",\"s3:ListBucketVersions\",
        \"s3:GetObjectVersion\",\"s3:DeleteObjectVersion\",
        \"s3:GetBucketVersioning\"
      ],
      \"Resource\": [
        \"arn:aws:s3:::${BUCKET_NAME}\",
        \"arn:aws:s3:::${BUCKET_NAME}/*\"
      ]
    }]
  }"

# 2. Create the IAM role for the instance profile
aws iam create-role \
  --role-name "sdk-s3-iam-backuper-role" \
  --assume-role-policy-document '{
    "Version":"2012-10-17",
    "Statement":[{"Effect":"Allow","Principal":{"Service":"ec2.amazonaws.com"},"Action":"sts:AssumeRole"}]
  }'

# 3. Attach the policy to the role
POLICY_ARN=$(aws iam list-policies --scope Local \
  --query "Policies[?PolicyName=='sdk-s3-iam-backuper-policy'].Arn" \
  --output text)
aws iam attach-role-policy \
  --role-name "sdk-s3-iam-backuper-role" \
  --policy-arn "${POLICY_ARN}"

# 4. Create the instance profile and add the role
aws iam create-instance-profile \
  --instance-profile-name "${PROFILE_NAME}"
aws iam add-role-to-instance-profile \
  --instance-profile-name "${PROFILE_NAME}" \
  --role-name "sdk-s3-iam-backuper-role"

echo "Instance profile '${PROFILE_NAME}' created."
echo "Set this as the value of the 'bbl_aws_iam_instance_profile' CredHub credential."
IAMEOF
fi

if [[ ${FAIL} -gt 0 ]]; then
  echo ""
  echo "  ACTION REQUIRED: ${FAIL} check(s) failed — see above."
  exit 1
else
  echo "  All required permissions confirmed."
  exit 0
fi
