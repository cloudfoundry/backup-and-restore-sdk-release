#!/usr/bin/env bash
#
# Creates the GCP resources the backup-and-restore-sdk-release pipeline needs
# in the GCP project cloud-foundry-310819.
#
#   bosh-bbr-sdk-pipeline     Service account with roles/storage.admin on the
#                             project. The contract tests create and delete
#                             their own contract-test-* buckets, so the role
#                             can't be limited to the buckets below.
#   ci-bbr-sdk-gcs-live       Bucket the gcs-backuper system tests back up.
#   ci-bbr-sdk-gcs-backup     Bucket they back up into.
#
# The buckets are unversioned: the GCS backuper doesn't use versioning.
# They're in europe-west2, where the contract tests create their buckets.
#
# Safe to re-run: the service account, buckets and role binding are kept if
# they already exist. A key is only created when the service account has no
# user-managed keys (set FORCE_NEW_KEY=true to create another anyway).
#
# Needs: gcloud, logged in as someone who can create service accounts and
# keys, create buckets, and set IAM policy on the project (e.g. roles/owner).
#
# Usage:
#   ./ci/scripts/create-gcp.sh
#   DRY_RUN=true ./ci/scripts/create-gcp.sh   # print commands only

set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

PROJECT_ID="cloud-foundry-310819"
REGION="europe-west2"
DRY_RUN="${DRY_RUN:-false}"
FORCE_NEW_KEY="${FORCE_NEW_KEY:-false}"

SA_NAME="bosh-bbr-sdk-pipeline"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
SA_ROLE="roles/storage.admin"

BUCKETS=(
  ci-bbr-sdk-gcs-live
  ci-bbr-sdk-gcs-backup
)

KEY_FILE="${KEY_FILE:-./bbr_sdk_gcp_service_account_key.json}"

run() {
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "+ $*"
  else
    "$@"
  fi
}

check_project() {
  if ! gcloud projects describe "${PROJECT_ID}" --format='value(projectId)' >/dev/null 2>&1; then
    echo "ERROR: can't see project ${PROJECT_ID} as $(gcloud config get-value account 2>/dev/null)" >&2
    exit 1
  fi
  echo "account: $(gcloud config get-value account 2>/dev/null)"
  echo "project: ${PROJECT_ID}"
}

ensure_service_account() {
  if gcloud iam service-accounts describe "${SA_EMAIL}" --project "${PROJECT_ID}" >/dev/null 2>&1; then
    echo "exists:  service account ${SA_EMAIL}"
  else
    echo "create:  service account ${SA_EMAIL}"
    run gcloud iam service-accounts create "${SA_NAME}" \
      --project "${PROJECT_ID}" \
      --display-name "backup-and-restore-sdk-release pipeline" \
      --description "GCS contract and system tests for the backup-and-restore-sdk-release pipeline"
  fi
}

ensure_role_binding() {
  echo "bind:    ${SA_ROLE} to ${SA_EMAIL}"
  # A just-created service account can take a few seconds to be accepted as
  # a policy member, so retry.
  local attempt
  for attempt in 1 2 3 4 5 6; do
    if run gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
      --member "serviceAccount:${SA_EMAIL}" \
      --role "${SA_ROLE}" \
      --condition None \
      --quiet >/dev/null; then
      return
    fi
    if [[ "${attempt}" == 6 ]]; then
      echo "ERROR: could not bind ${SA_ROLE} to ${SA_EMAIL}" >&2
      exit 1
    fi
    sleep 10
  done
}

ensure_bucket() {
  local bucket="$1"

  if gcloud storage buckets describe "gs://${bucket}" --project "${PROJECT_ID}" >/dev/null 2>&1; then
    echo "exists:  bucket gs://${bucket}"
  else
    echo "create:  bucket gs://${bucket} (${REGION})"
    run gcloud storage buckets create "gs://${bucket}" \
      --project "${PROJECT_ID}" \
      --location "${REGION}" \
      --uniform-bucket-level-access \
      --public-access-prevention
  fi
}

ensure_key() {
  local keys
  keys="$(gcloud iam service-accounts keys list \
    --iam-account "${SA_EMAIL}" \
    --project "${PROJECT_ID}" \
    --managed-by user \
    --format 'value(name.basename())' 2>/dev/null || true)"
  if [[ -n "${keys}" && "${FORCE_NEW_KEY}" != "true" ]]; then
    echo "exists:  key(s) for ${SA_EMAIL}: $(echo "${keys}" | tr '\n' ' ')(not creating another)"
    echo "         Set FORCE_NEW_KEY=true to create one anyway."
    return
  fi

  echo "create:  key for ${SA_EMAIL} in ${KEY_FILE}"
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "+ gcloud iam service-accounts keys create ${KEY_FILE} --iam-account ${SA_EMAIL}"
    return
  fi

  if [[ -e "${KEY_FILE}" ]]; then
    echo "ERROR: ${KEY_FILE} already exists; move it away or set KEY_FILE" >&2
    exit 1
  fi

  (
    umask 077
    gcloud iam service-accounts keys create "${KEY_FILE}" \
      --iam-account "${SA_EMAIL}" \
      --project "${PROJECT_ID}"
  ) || {
    echo "ERROR: key creation failed. If the error mentions" >&2
    echo "  constraints/iam.disableServiceAccountKeyCreation, an org policy blocks" >&2
    echo "  keys in this project; ask an org admin for an exception." >&2
    exit 1
  }

  cat <<EOF

The key is in ${KEY_FILE}. It can't be downloaded again.

Send it to the pipeline maintainers over a secure channel (not email or
chat), then delete your copy. They store it in CredHub with:

  credhub set -t json \\
    -n /concourse/main/backup-and-restore-sdk-release/bbr_sdk_gcp_service_account_key \\
    -v "\$(cat ${KEY_FILE})"

Values the pipeline uses:
  gcp_project_name:       ${PROJECT_ID}
  gcs_bucket_name:        ${BUCKETS[0]}
  gcs_backup_bucket_name: ${BUCKETS[1]}
EOF
}

main() {
  check_project

  ensure_service_account
  ensure_role_binding
  local bucket
  for bucket in "${BUCKETS[@]}"; do
    ensure_bucket "${bucket}"
  done
  ensure_key
}

main "$@"
