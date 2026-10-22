#!/usr/bin/env bash
#
# Creates the S3 buckets the backup-and-restore-sdk-release pipeline needs in
# the AWS account 504405067906.
#
# Primary region is eu-north-1. Buckets that tests expect to be in a
# *different* region from their source (clone / backup buckets, and the
# contract tests' backup region) go in eu-west-1.
#
# Safe to re-run: buckets that already exist in this account are skipped,
# and versioning / public-access settings are (re)applied.
#
# Usage:
#   ./ci/scripts/create-aws-s3-buckets.sh            # create buckets
#   DRY_RUN=true ./ci/scripts/create-aws-s3-buckets.sh  # print commands only

set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

EXPECTED_ACCOUNT_ID="504405067906"
PRIMARY_REGION="eu-north-1"
SECONDARY_REGION="eu-west-1"
DRY_RUN="${DRY_RUN:-false}"

# name:region:versioned
BUCKETS=(
  # Pipeline state: version, dev-release, RDS terraform state
  "ci-bbr-sdk-pipeline:${PRIMARY_REGION}:true"

  # Contract-test fixtures (pre-seeded with big_file)
  "ci-bbr-sdk-large-blob:${PRIMARY_REGION}:true"
  "ci-bbr-sdk-large-blob-unversioned:${PRIMARY_REGION}:false"
  # Contract-test fixtures (pre-seeded: exactly 2001 object versions / 2001 objects)
  "ci-bbr-sdk-big-bucket:${PRIMARY_REGION}:true"
  "ci-bbr-sdk-unversioned-big-bucket:${PRIMARY_REGION}:false"

  # s3-backuper system tests: versioned
  "ci-bbr-sdk-system-test:${PRIMARY_REGION}:true"
  "ci-bbr-sdk-system-test-clone:${SECONDARY_REGION}:true"
  # Must stay unversioned: negative test for the versioned backuper
  "ci-bbr-sdk-system-test-unversioned:${PRIMARY_REGION}:false"

  # s3-backuper system tests: unversioned
  "ci-bbr-sdk-s3-unversioned:${PRIMARY_REGION}:false"
  "ci-bbr-sdk-s3-unversioned-backup:${SECONDARY_REGION}:false"
  "ci-bbr-sdk-unversioned-bpm:${PRIMARY_REGION}:false"
  "ci-bbr-sdk-unversioned-bpm-backup:${PRIMARY_REGION}:false"
  "ci-bbr-sdk-large-number-of-files:${PRIMARY_REGION}:false"
  "ci-bbr-sdk-large-number-of-files-backup:${PRIMARY_REGION}:false"
  "ci-bbr-sdk-unversioned-clone:${SECONDARY_REGION}:false"

  # IAM instance-profile system test
  "ci-bbr-sdk-iam-instance-role-test:${PRIMARY_REGION}:true"
  "ci-bbr-sdk-iam-instance-role-test-clone:${PRIMARY_REGION}:true"
)

run() {
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "+ $*"
  else
    "$@"
  fi
}

bucket_exists() {
  aws s3api head-bucket --bucket "$1" >/dev/null 2>&1
}

create_bucket() {
  local bucket="$1" region="$2"

  if bucket_exists "${bucket}"; then
    echo "exists:  ${bucket}"
    return
  fi

  echo "create:  ${bucket} (${region})"
  run aws s3api create-bucket \
    --bucket "${bucket}" \
    --region "${region}" \
    --create-bucket-configuration "LocationConstraint=${region}" \
    --query Location --output text
}

configure_bucket() {
  local bucket="$1" region="$2" versioned="$3"

  run aws s3api put-public-access-block \
    --bucket "${bucket}" \
    --region "${region}" \
    --public-access-block-configuration \
      "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"

  if [[ "${versioned}" == "true" ]]; then
    run aws s3api put-bucket-versioning \
      --bucket "${bucket}" \
      --region "${region}" \
      --versioning-configuration Status=Enabled
  fi
}

check_account() {
  local account_id
  account_id="$(aws sts get-caller-identity --query Account --output text)"
  if [[ "${account_id}" != "${EXPECTED_ACCOUNT_ID}" ]]; then
    echo "ERROR: logged in to AWS account ${account_id}, expected ${EXPECTED_ACCOUNT_ID}" >&2
    exit 1
  fi
}

summarize() {
  echo ""
  printf '%-42s %-12s %s\n' "BUCKET" "REGION" "VERSIONING"
  for entry in "${BUCKETS[@]}"; do
    IFS=: read -r bucket _ _ <<< "${entry}"
    printf '%-42s %-12s %s\n' "${bucket}" \
      "$(aws s3api get-bucket-location --bucket "${bucket}" --query LocationConstraint --output text)" \
      "$(aws s3api get-bucket-versioning --bucket "${bucket}" --query Status --output text)"
  done
}

main() {
  check_account

  for entry in "${BUCKETS[@]}"; do
    IFS=: read -r bucket region versioned <<< "${entry}"
    create_bucket "${bucket}" "${region}"
    configure_bucket "${bucket}" "${region}" "${versioned}"
  done

  if [[ "${DRY_RUN}" != "true" ]]; then
    summarize
  fi
}

main "$@"
