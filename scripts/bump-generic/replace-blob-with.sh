#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
SDK_ROOT="$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel)"

CUR_BLOB_ID="${1}"
NEW_BLOB_ID="${2}"
CUR_VERSION="${3}"
NEW_VERSION="${4}"
NEW_TARFILE="${5}"

# TODO: Extend this check to also compare shasum and id
if [[ "${CUR_VERSION}" == "${NEW_VERSION}" ]];
then
  echo "${CUR_BLOB_ID} is up-to-date"
  exit 0
fi

pushd "${SDK_ROOT}" >/dev/null

echo "Replacing Postgresql ${CUR_VERSION} with ${NEW_VERSION}"

# Replace blobstore blob
bosh remove-blob "--dir=${SDK_ROOT}" "${CUR_BLOB_ID}"
bosh add-blob "--dir=${SDK_ROOT}" "$(realpath "${NEW_TARFILE}")" "${NEW_BLOB_ID}"
bosh upload-blobs "--dir=${SDK_ROOT}"

# Replace references in files
# Following steps **SHOULD NEVER BE PERFORMED BEFORE** the blobstore replacement commands listed above
FILES_WITH_REFS="$(grep -rnwl '.' -e "${CUR_BLOB_ID}")"

for file in $FILES_WITH_REFS;
do
  sed -i '' "s#${CUR_BLOB_ID}#${NEW_BLOB_ID}#g" "$file"
  sed -i '' "s#${CUR_VERSION}#${NEW_VERSION}#g" "$file"
done

popd >/dev/null
