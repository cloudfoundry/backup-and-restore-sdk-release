#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
SDK_ROOT="$(git rev-parse --show-toplevel)"

function download_version() {
  VERSION="$1"
  OUTFILE="mariadb-${VERSION}.tar.gz"
  wget -q -O "${OUTFILE}" "https://downloads.mariadb.org/interstitial/mariadb-${VERSION}/source/mariadb-${VERSION}.tar.gz"
  echo "${OUTFILE}"
}

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
  NEW_BLOB_VERSION="$2"
  NEW_BLOB_FILE=$(download_version "${NEW_BLOB_VERSION}")

  bosh remove-blob "--dir=${SDK_ROOT}" "mariadb/mariadb-${CUR_BLOB_VERSION}.tar.gz"
  bosh add-blob "--dir=${SDK_ROOT}" "${SDK_ROOT}/${NEW_BLOB_FILE}" "mariadb/mariadb-${NEW_BLOB_VERSION}.tar.gz"
  bosh upload-blobs "--dir=${SDK_ROOT}"
  rm "${NEW_BLOB_FILE}"
}

function replace_references_in_files() {
  NEW_BLOB_VERSION="$1"

  REPLACED="$(sed -e "s/mariadb-.*.tar.gz/mariadb-${NEW_BLOB_VERSION}.tar.gz/" \
      "${SDK_ROOT}/packages/database-backup-restorer-mariadb/spec")"
  echo "${REPLACED}" > "${SDK_ROOT}/packages/database-backup-restorer-mariadb/spec"

  REPLACED="$(sed -e "s/MARIADB_VERSION=.*$/MARIADB_VERSION=${NEW_BLOB_VERSION}/" \
      "${SDK_ROOT}/packages/database-backup-restorer-mariadb/packaging")"
  echo "${REPLACED}" > "${SDK_ROOT}/packages/database-backup-restorer-mariadb/packaging"
}

function ensure_blobstoreid_exists() {
  BLOBSTORE_ID="$(bosh blobs | grep "mariadb" | cut -f3 | xargs)"

  if [[ "${BLOBSTORE_ID}" == "(local)" ]];
  then
    echo "But its Blobstore ID was not found. Uploading..."
    bosh upload-blobs "--dir=${SDK_ROOT}"
  fi
}

VERSION="${1}"

pushd "${SCRIPT_DIR}" >/dev/null
CURRENT="$(./current_version.sh)"
popd >/dev/null

pushd "${SDK_ROOT}" >/dev/null

if [[ "${CURRENT}" == "${VERSION}" ]];
then
  echo "Already at version ${VERSION}"
  ensure_blobstoreid_exists "${CURRENT}"
elif [[ "$(get_newest_version "${CURRENT}" "${VERSION}")" == "${CURRENT}" ]];
then
  echo "Current version '${CURRENT}' is more recent than '${VERSION}'"
  ensure_blobstoreid_exists "${CURRENT}"
else
  echo "Updating MariaDB from ${CURRENT} to ${VERSION}"
  replace_references_in_files "${VERSION}"
  replace_blobstore_version "${CURRENT}" "${VERSION}"
fi
popd >/dev/null
