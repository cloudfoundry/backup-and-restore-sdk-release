#!/usr/bin/env bash
#
# Creates the IAM user and role that write to the release blobstore in the
# AWS account 504405067906, and stores the credentials in CredHub.
#
#   bosh_bbr_sdk_release_blobstore        IAM user. Its only permission is
#                                         assuming the role below.
#   bosh_bbr_sdk_release_blobstore_role   Role that writes to the release
#                                         blobstore. Used by
#                                         create-final-release and
#                                         bump-packages (bosh upload-blobs,
#                                         via private.yml assume_role_arn),
#                                         and by the release-notes resource.
#
# These are separate from bbr_sdk_aws_creds so the test credentials, which
# can create EC2, RDS and IAM resources, can't touch the final release blobs.
# The role can read and add objects but not delete them: final release blobs
# are never removed, and versioning keeps overwritten release-notes.md.
#
# The access key goes straight into CredHub as
#   /concourse/main/backup-and-restore-sdk-release/bbr_sdk_release_blobstore_creds
# (access_key_id, secret_access_key, role_arn), so the secret is never printed.
#
# Safe to re-run: the user and role are kept, and policies are (re)applied.
# The access key is only created when the user has none.
#
# Usage:
#   ./ci/scripts/create-release-blobstore-iam.sh
#   DRY_RUN=true ./ci/scripts/create-release-blobstore-iam.sh   # print commands only

set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

ACCOUNT_ID="504405067906"
BUCKET="backup-and-restore-sdk-release-blobs"
DRY_RUN="${DRY_RUN:-false}"

USER_NAME="bosh_bbr_sdk_release_blobstore"
ROLE_NAME="bosh_bbr_sdk_release_blobstore_role"

CREDHUB_API="https://10.0.0.4:8844"
CREDHUB_NAME="/concourse/main/backup-and-restore-sdk-release/bbr_sdk_release_blobstore_creds"

run() {
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "+ $*"
  else
    "$@"
  fi
}

check_account() {
  local account_id
  account_id="$(aws sts get-caller-identity --query Account --output text)"
  if [[ "${account_id}" != "${ACCOUNT_ID}" ]]; then
    echo "ERROR: logged in to AWS account ${account_id}, expected ${ACCOUNT_ID}" >&2
    exit 1
  fi
}

# Checked before any IAM changes, so a new access key is never created when
# it can't be stored.
check_credhub() {
  local target
  target="$(credhub api 2>/dev/null || true)"
  if [[ "${target}" != *"${CREDHUB_API}"* ]]; then
    echo "ERROR: credhub CLI isn't targeting ${CREDHUB_API}" >&2
    echo "${target}" >&2
    exit 1
  fi
  if ! credhub find -p /concourse/main/backup-and-restore-sdk-release >/dev/null 2>&1; then
    echo "ERROR: can't read CredHub. Log in first." >&2
    exit 1
  fi
}

user_policy() {
  cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AssumeReleaseBlobstoreRole",
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Resource": "arn:aws:iam::${ACCOUNT_ID}:role/${ROLE_NAME}"
    }
  ]
}
EOF
}

role_trust_policy() {
  cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "AWS": "arn:aws:iam::${ACCOUNT_ID}:user/${USER_NAME}" },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

role_policy() {
  cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ListReleaseBlobstore",
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket",
        "s3:ListBucketVersions",
        "s3:GetBucketLocation",
        "s3:GetBucketVersioning"
      ],
      "Resource": "arn:aws:s3:::${BUCKET}"
    },
    {
      "Sid": "ReadWriteReleaseBlobs",
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:PutObject"
      ],
      "Resource": "arn:aws:s3:::${BUCKET}/*"
    }
  ]
}
EOF
}

ensure_user() {
  if aws iam get-user --user-name "${USER_NAME}" >/dev/null 2>&1; then
    echo "exists:  user ${USER_NAME}"
  else
    echo "create:  user ${USER_NAME}"
    run aws iam create-user --user-name "${USER_NAME}" >/dev/null
  fi
  run aws iam put-user-policy \
    --user-name "${USER_NAME}" \
    --policy-name assume-release-blobstore-role \
    --policy-document "$(user_policy)"
}

ensure_role() {
  if aws iam get-role --role-name "${ROLE_NAME}" >/dev/null 2>&1; then
    echo "exists:  role ${ROLE_NAME}"
    run aws iam update-assume-role-policy --role-name "${ROLE_NAME}" --policy-document "$(role_trust_policy)"
  else
    echo "create:  role ${ROLE_NAME}"
    # A just-created user can take a few seconds to be accepted as a trust
    # policy principal ("Invalid principal in policy"), so retry.
    local attempt
    for attempt in 1 2 3 4 5 6; do
      if run aws iam create-role \
        --role-name "${ROLE_NAME}" \
        --assume-role-policy-document "$(role_trust_policy)" \
        --query Role.Arn --output text; then
        break
      fi
      if [[ "${attempt}" == 6 ]]; then
        echo "ERROR: could not create role ${ROLE_NAME}" >&2
        exit 1
      fi
      sleep 10
    done
  fi
  run aws iam put-role-policy \
    --role-name "${ROLE_NAME}" \
    --policy-name release-blobstore \
    --policy-document "$(role_policy)"
}

ensure_access_key() {
  local keys
  keys="$(aws iam list-access-keys --user-name "${USER_NAME}" \
    --query 'AccessKeyMetadata[].AccessKeyId' --output text 2>/dev/null || true)"
  if [[ -n "${keys}" && "${keys}" != "None" ]]; then
    echo "exists:  access key(s) for ${USER_NAME}: ${keys} (not creating another)"
    return
  fi

  echo "create:  access key for ${USER_NAME}"
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "+ aws iam create-access-key --user-name ${USER_NAME}"
    echo "+ credhub set -t json -n ${CREDHUB_NAME} -v <access key>"
    return
  fi

  local key key_id secret
  key="$(aws iam create-access-key --user-name "${USER_NAME}" \
    --query 'AccessKey.[AccessKeyId,SecretAccessKey]' --output text)"
  read -r key_id secret <<< "${key}"

  echo "store:   ${CREDHUB_NAME}"
  # If the key can't be stored, delete it, so a re-run creates a fresh one
  # rather than finding a key nobody has the secret for.
  if ! credhub set -t json -n "${CREDHUB_NAME}" -v "$(cat <<EOF
{
  "access_key_id": "${key_id}",
  "secret_access_key": "${secret}",
  "role_arn": "arn:aws:iam::${ACCOUNT_ID}:role/${ROLE_NAME}"
}
EOF
)" >/dev/null; then
    echo "ERROR: credhub set failed, deleting access key ${key_id}" >&2
    aws iam delete-access-key --user-name "${USER_NAME}" --access-key-id "${key_id}"
    exit 1
  fi
}

main() {
  check_account
  check_credhub

  ensure_user
  ensure_role
  ensure_access_key
}

main "$@"
