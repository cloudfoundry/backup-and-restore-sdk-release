#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
SDK_ROOT="$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel)"

function get_newest_version() {
  VERSION1="${1}"
  VERSION2="${2}"
  NEWEST="$(echo "${VERSION1}"$'\n'"${VERSION2}" | sort -t "." -k1,1n -k2,2n -k3,3n | tail -n 1)"

  if [[ "${NEWEST}" == "${VERSION1}" ]];
  then
    echo "${VERSION1}"
  elif [[ "${NEWEST}" == "${VERSION2}" ]];
  then
    echo "${VERSION2}"
  else
    echo "Impossible state reached. This is a bug."
    exit 1
  fi
}

function replace_blobstore_version() {
  CUR_BLOB_VERSION="$1"
  NEW_BLOB_FILEPATH="$2"
  NEW_BLOB_FILENAME="$(basename "${NEW_BLOB_FILEPATH}")"

  bosh remove-blob "--dir=${SDK_ROOT}" "mysql/mysql-${CUR_BLOB_VERSION}.tar.gz"
  bosh add-blob "--dir=${SDK_ROOT}" "${NEW_BLOB_FILEPATH}" "mysql/${NEW_BLOB_FILENAME}"
  bosh upload-blobs "--dir=${SDK_ROOT}"
}

function replace_references_in_files() {
  NEW_BLOB_VERSION="$1"

  REPLACED="$(sed -e "s/mysql-.*.tar.gz/mysql-${NEW_BLOB_VERSION}.tar.gz/" \
      "${SDK_ROOT}/packages/database-backup-restorer-mysql-${MAJOR_MINOR}/spec")"
  echo "${REPLACED}" > "${SDK_ROOT}/packages/database-backup-restorer-mysql-${MAJOR_MINOR}/spec"

  REPLACED="$(sed -e "s/MYSQL_VERSION=.*$/MYSQL_VERSION=${NEW_BLOB_VERSION}/" \
      "${SDK_ROOT}/packages/database-backup-restorer-mysql-${MAJOR_MINOR}/packaging")"
  echo "${REPLACED}" > "${SDK_ROOT}/packages/database-backup-restorer-mysql-${MAJOR_MINOR}/packaging"
}

function ensure_blobstoreid_exists() {
  BLOBSTORE_ID="$(bosh blobs | grep "mysql" | cut -f3 | xargs)"

  if [[ "${BLOBSTORE_ID}" == "(local)" ]];
  then
    echo "But its Blobstore ID was not found. Uploading..."
    bosh upload-blobs "--dir=${SDK_ROOT}"
  fi
}

NEW_BLOB_FILE="${1}"
NEW_VERSION="$(basename "${NEW_BLOB_FILE}"| sed -n 's/^mysql-\([0-9]*\.[0-9]*\.[0-9]*\)\.tar\.gz$/\1/p')"
NEW_BLOB_ABS_PATH="$(realpath "${NEW_BLOB_FILE}")"

if [[ -z "${NEW_VERSION}" ]];
then
  echo "Provided file ${NEW_BLOB_FILE} doesn't match the required naming convention:"
  echo "mysql-{SEMVER_NUMBER}.tar.gz"
  exit 1
fi

pushd "${SCRIPT_DIR}" >/dev/null
MAJOR_MINOR="$(echo "${NEW_VERSION}" | cut -d '.' -f1,2)"
CURRENT="$(./current_blobs_version.sh | grep "${MAJOR_MINOR}")"
popd >/dev/null

pushd "${SDK_ROOT}" >/dev/null

if [[ "${CURRENT}" == "${NEW_VERSION}" ]];
then
  echo "Already at version ${NEW_VERSION}"
  ensure_blobstoreid_exists "${CURRENT}"
elif [[ "$(get_newest_version "${CURRENT}" "${NEW_VERSION}")" == "${CURRENT}" ]];
then
  echo "Current version '${CURRENT}' is more recent than '${NEW_VERSION}'"
  ensure_blobstoreid_exists "${CURRENT}"
else
  echo "Updating MySQL from ${CURRENT} to ${NEW_VERSION}"
  replace_references_in_files "${NEW_VERSION}"
  replace_blobstore_version "${CURRENT}" "${NEW_BLOB_ABS_PATH}"
fi
popd >/dev/null
