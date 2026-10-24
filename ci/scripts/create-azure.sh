#!/usr/bin/env bash
#
# Creates the Azure resources the backup-and-restore-sdk-release pipeline
# needs in an Azure subscription.
#
#   resource group ci-bbr-sdk
#   storage account 1             Blob soft delete enabled. The backuper refuses
#                                 accounts without it (backuper.go).
#     bbr-system-test-azure-container
#     bbr-system-test-azure-different-container   restore target for the
#                                           same-account "different
#                                           container" system test
#     bbr-test-many-blobs-azure-container   seeded with exactly 10104 blobs
#                                           (contract_test/container_test.go)
#   storage account 2             Soft delete enabled. Used for cross-account
#                                 restores.
#     bbr-system-test-azure-different-container
#   no-soft-delete account        Soft delete disabled, for the negative test.
#
# Blob versioning stays off on every account: restores copy the soft-delete
# snapshots of overwritten blobs, which Azure doesn't create when versioning
# is on. Shared key access stays on: the tests and the backuper authenticate
# with account keys, and cross-account copies use shared-key SAS URLs.
#
# The contract tests create and delete their own sdk-azure-test-* containers,
# so nothing else needs pre-creating.
#
# Safe to re-run: existing resources are kept and their settings (re)applied.
# The many-blobs container is only seeded when it's empty, and the script
# stops if it holds anything other than 10104 blobs.
#
# Needs: az, logged in (az login) as someone who can create resource groups
# and storage accounts in the subscription (e.g. Contributor).
#
# Usage:
#   AZURE_SUBSCRIPTION_ID=<id> ./ci/scripts/create-azure.sh
#   DRY_RUN=true AZURE_SUBSCRIPTION_ID=<id> ./ci/scripts/create-azure.sh   # print commands only
#
# Storage account names are global across Azure. If one is taken, override it
# with STORAGE_ACCOUNT_1, STORAGE_ACCOUNT_2 or STORAGE_ACCOUNT_NO_SOFT_DELETE
# (3-24 lowercase letters and digits).

set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

SUBSCRIPTION_ID="${AZURE_SUBSCRIPTION_ID:?set AZURE_SUBSCRIPTION_ID}"
RESOURCE_GROUP="${RESOURCE_GROUP:-ci-bbr-sdk}"
LOCATION="${LOCATION:-westeurope}"
DRY_RUN="${DRY_RUN:-false}"

STORAGE_ACCOUNT_1="${STORAGE_ACCOUNT_1:-cibbrsdk1}"
STORAGE_ACCOUNT_2="${STORAGE_ACCOUNT_2:-cibbrsdk2}"
STORAGE_ACCOUNT_NO_SOFT_DELETE="${STORAGE_ACCOUNT_NO_SOFT_DELETE:-cibbrsdknosoftdelete}"

SOFT_DELETE_DAYS=7

CONTAINER_1="bbr-system-test-azure-container"
CONTAINER_MANY_BLOBS="bbr-test-many-blobs-azure-container"
CONTAINER_2="bbr-system-test-azure-different-container"
MANY_BLOBS_COUNT=10104

CREDS_FILE="${CREDS_FILE:-./bbr_sdk_azure_credentials.json}"

run() {
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "+ $*"
  else
    "$@"
  fi
}

check_subscription() {
  run az account set --subscription "${SUBSCRIPTION_ID}"
  local name
  if ! name="$(az account show --subscription "${SUBSCRIPTION_ID}" --query name --output tsv 2>/dev/null)"; then
    echo "ERROR: can't see subscription ${SUBSCRIPTION_ID}; run az login first" >&2
    exit 1
  fi
  echo "account:      $(az account show --query user.name --output tsv)"
  echo "subscription: ${name} (${SUBSCRIPTION_ID})"
}

ensure_resource_group() {
  if [[ "$(az group exists --name "${RESOURCE_GROUP}" --subscription "${SUBSCRIPTION_ID}")" == "true" ]]; then
    echo "exists:  resource group ${RESOURCE_GROUP}"
  else
    echo "create:  resource group ${RESOURCE_GROUP} (${LOCATION})"
    run az group create \
      --name "${RESOURCE_GROUP}" \
      --location "${LOCATION}" \
      --subscription "${SUBSCRIPTION_ID}" \
      --only-show-errors >/dev/null
  fi
}

ensure_storage_account() {
  local account="$1" soft_delete="$2"

  if az storage account show --name "${account}" --resource-group "${RESOURCE_GROUP}" \
    --subscription "${SUBSCRIPTION_ID}" >/dev/null 2>&1; then
    echo "exists:  storage account ${account}"
  else
    local available
    available="$(az storage account check-name --name "${account}" --query nameAvailable --output tsv)"
    if [[ "${available}" != "true" ]]; then
      echo "ERROR: storage account name ${account} is taken outside ${RESOURCE_GROUP}; pick another (see Usage)" >&2
      exit 1
    fi

    echo "create:  storage account ${account} (${LOCATION})"
    run az storage account create \
      --name "${account}" \
      --resource-group "${RESOURCE_GROUP}" \
      --subscription "${SUBSCRIPTION_ID}" \
      --location "${LOCATION}" \
      --kind StorageV2 \
      --sku Standard_LRS \
      --min-tls-version TLS1_2 \
      --allow-blob-public-access false \
      --allow-shared-key-access true \
      --only-show-errors >/dev/null
  fi

  echo "set:     ${account} blob soft delete=${soft_delete}, versioning=false"
  local retention=(--enable-delete-retention false)
  if [[ "${soft_delete}" == "true" ]]; then
    retention=(--enable-delete-retention true --delete-retention-days "${SOFT_DELETE_DAYS}")
  fi
  run az storage account blob-service-properties update \
    --account-name "${account}" \
    --resource-group "${RESOURCE_GROUP}" \
    --subscription "${SUBSCRIPTION_ID}" \
    "${retention[@]}" \
    --enable-versioning false \
    --only-show-errors >/dev/null
}

account_key() {
  local account="$1"
  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "<${account}-key>"
    return
  fi
  az storage account keys list \
    --account-name "${account}" \
    --resource-group "${RESOURCE_GROUP}" \
    --subscription "${SUBSCRIPTION_ID}" \
    --query '[0].value' --output tsv
}

ensure_container() {
  local account="$1" container="$2"
  local key
  key="$(account_key "${account}")"

  if [[ "${DRY_RUN}" != "true" ]] && [[ "$(az storage container exists \
    --name "${container}" \
    --account-name "${account}" \
    --account-key "${key}" \
    --query exists --output tsv)" == "true" ]]; then
    echo "exists:  container ${account}/${container}"
  else
    echo "create:  container ${account}/${container}"
    run az storage container create \
      --name "${container}" \
      --account-name "${account}" \
      --account-key "${key}" \
      --only-show-errors >/dev/null
  fi
}

seed_many_blobs() {
  local account="$1" container="$2"
  local key
  key="$(account_key "${account}")"

  if [[ "${DRY_RUN}" == "true" ]]; then
    echo "seed:    ${account}/${container} with ${MANY_BLOBS_COUNT} blobs"
    echo "+ az storage blob upload-batch --destination ${container} --account-name ${account} ..."
    return
  fi

  local count
  count="$(az storage blob list \
    --container-name "${container}" \
    --account-name "${account}" \
    --account-key "${key}" \
    --num-results '*' \
    --query 'length(@)' --output tsv \
    --only-show-errors)"

  if [[ "${count}" == "${MANY_BLOBS_COUNT}" ]]; then
    echo "exists:  ${account}/${container} has ${MANY_BLOBS_COUNT} blobs"
    return
  fi
  if [[ "${count}" != "0" ]]; then
    echo "ERROR: ${account}/${container} has ${count} blobs, expected 0 or ${MANY_BLOBS_COUNT}." >&2
    echo "  Empty it and re-run: az storage blob delete-batch --source ${container} --account-name ${account}" >&2
    exit 1
  fi

  echo "seed:    ${account}/${container} with ${MANY_BLOBS_COUNT} blobs (this takes a while)"
  SEED_DIR="$(mktemp -d)"
  trap 'rm -rf "${SEED_DIR}"' EXIT
  local dir="${SEED_DIR}"

  local i name
  for ((i = 1; i <= MANY_BLOBS_COUNT; i++)); do
    printf -v name 'blob-%05d' "${i}"
    printf 'blob %d\n' "${i}" > "${dir}/${name}"
  done

  az storage blob upload-batch \
    --destination "${container}" \
    --source "${dir}" \
    --account-name "${account}" \
    --account-key "${key}" \
    --only-show-errors >/dev/null
  rm -rf "${dir}"

  count="$(az storage blob list \
    --container-name "${container}" \
    --account-name "${account}" \
    --account-key "${key}" \
    --num-results '*' \
    --query 'length(@)' --output tsv \
    --only-show-errors)"
  if [[ "${count}" != "${MANY_BLOBS_COUNT}" ]]; then
    echo "ERROR: ${account}/${container} has ${count} blobs after seeding, expected ${MANY_BLOBS_COUNT}" >&2
    exit 1
  fi
}

write_credentials() {
  echo "write:   ${CREDS_FILE}"
  if [[ "${DRY_RUN}" == "true" ]]; then
    return
  fi

  if [[ -e "${CREDS_FILE}" ]]; then
    echo "ERROR: ${CREDS_FILE} already exists; move it away or set CREDS_FILE" >&2
    exit 1
  fi

  (
    umask 077
    cat > "${CREDS_FILE}" <<EOF
{
  "storage_account": "${STORAGE_ACCOUNT_1}",
  "storage_account_key": "$(account_key "${STORAGE_ACCOUNT_1}")",
  "storage_account2": "${STORAGE_ACCOUNT_2}",
  "storage_account2_key": "$(account_key "${STORAGE_ACCOUNT_2}")",
  "storage_account_no_soft_delete": "${STORAGE_ACCOUNT_NO_SOFT_DELETE}",
  "storage_account_no_soft_delete_key": "$(account_key "${STORAGE_ACCOUNT_NO_SOFT_DELETE}")"
}
EOF
  )

  cat <<EOF

The account names and keys are in ${CREDS_FILE}. If it's lost, delete it
and re-run this script to regenerate it from the current keys.

Send it to the pipeline maintainers over a secure channel (not email or
chat), then delete your copy. They store it in CredHub with:

  credhub set -t json \\
    -n /concourse/main/backup-and-restore-sdk-release/bbr_sdk_azure_credentials \\
    -v "\$(cat ${CREDS_FILE})"

Values the pipeline uses:
  azure_container_name:           ${CONTAINER_1}
  azure_different_container_name: ${CONTAINER_2}
  azure_container_name_many_files: ${CONTAINER_MANY_BLOBS}
EOF
}

main() {
  check_subscription

  ensure_resource_group
  ensure_storage_account "${STORAGE_ACCOUNT_1}" true
  ensure_storage_account "${STORAGE_ACCOUNT_2}" true
  ensure_storage_account "${STORAGE_ACCOUNT_NO_SOFT_DELETE}" false

  ensure_container "${STORAGE_ACCOUNT_1}" "${CONTAINER_1}"
  ensure_container "${STORAGE_ACCOUNT_1}" "${CONTAINER_2}"
  ensure_container "${STORAGE_ACCOUNT_1}" "${CONTAINER_MANY_BLOBS}"
  ensure_container "${STORAGE_ACCOUNT_2}" "${CONTAINER_2}"

  seed_many_blobs "${STORAGE_ACCOUNT_1}" "${CONTAINER_MANY_BLOBS}"

  write_credentials
}

main "$@"
