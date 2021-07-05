#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
SDK_ROOT="$(cd "${SCRIPT_DIR}/.." &>/dev/null && pwd)"

pushd "${SDK_ROOT}" >/dev/null

function check_requirements() {
  if ! command -v bosh &> /dev/null; then
      echo "bosh could not be found"; exit
  fi
  if ! command -v xmllint &> /dev/null; then
      echo "xmllint could not be found"; exit
  fi
  if ! command -v wget &> /dev/null; then
      echo "wget could not be found"; exit
  fi
  if ! command -v sed &> /dev/null; then
      echo "sed could not be found"; exit
  fi
}

function current_blob_name() {
  bosh blobs | grep 'mariadb' | cut -f1 | xargs
}

function current_blob_version() {
  current_blob_name | sed -n 's/^.*mariadb-\(.*\).tar.gz.*$/\1/p'
}

function stable_releases() {
  HTML="$(curl -s -L https://downloads.mariadb.org/mariadb/+releases/)"
  HREFS="$(echo "${HTML}" | xmllint --html --xpath "//table[@id='download']/tbody/tr[td[3]='Stable']/td[1]/a/@href" - 2>/dev/null)"
  VERSIONS="$(echo "${HREFS}" | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+')"
  echo "${VERSIONS}"
}

function last_major_release() {
  stable_releases | sort -t "." -k1,1n -k2,2n -k3,3n | tail -n 1
}

function last_patch_release() {
  CURRENT="$(current_blob_version | cut -d '.' -f1,2)"
  stable_releases | (grep "^${CURRENT}" || true) | sort -t "." -k1,1n -k2,2n -k3,3n | tail -n 1
}

function last_minor_release() {
  CURRENT="$(current_blob_version | cut -d '.' -f1)"
  stable_releases | (grep "^${CURRENT}" || true) | sort -t "." -k1,1n -k2,2n -k3,3n | tail -n 1
}

function download_version() {
  VERSION="$1"
  OUTFILE="mariadb-${VERSION}.tar.gz"
  wget -q -O "${OUTFILE}" "https://downloads.mariadb.org/interstitial/mariadb-${VERSION}/source/mariadb-${VERSION}.tar.gz"
  echo "${OUTFILE}"
}

function current_blob_newer_than() {
  VERSION="${1}"
  CURRENT="$(current_blob_version)"
  NEWEST="$(echo "${VERSION}"$'\n'"${CURRENT}" | sort -t "." -k1,1n -k2,2n -k3,3n | tail -n 1)"

  if [[ "${NEWEST}" == "${CURRENT}" ]];
  then
    echo "TRUE"
  elif [[ "${NEWEST}" == "${VERSION}" ]];
  then
    echo "FALSE"
  else
    echo "Impossible state reached. This is a bug."
    exit 1
  fi
}

function replace_blobstore_version() {
  NEW_BLOB_VERSION="$1"
  NEW_BLOB_FILE=$(download_version "${NEW_BLOB_VERSION}")

  bosh remove-blob "--dir=${SDK_ROOT}" "$(current_blob_name)"
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
    replace_blobstore_version "$(current_blob_version)"
  fi
}

function fetch_requested_semver_level() {
  SEMVER_LEVEL="$1"

  if [[ "${SEMVER_LEVEL}" == "patch" ]]; then
    VERSION="$(last_patch_release)"
  elif [[ "${SEMVER_LEVEL}" == "minor" ]]; then
    VERSION="$(last_minor_release)"
  elif [[ "${SEMVER_LEVEL}" == "major" ]]; then
    VERSION="$(last_major_release)"
  else
    >&2 echo "'${SEMVER_LEVEL}' is not a valid option: 'patch' 'minor' 'major'"
    exit 1
  fi
  echo "${VERSION}"
}

check_requirements
SEMVER_LEVEL="${1:-patch}" # 'patch' 'minor' 'major'
echo "Checking latest ${SEMVER_LEVEL} release"

CURRENT="$(current_blob_version)"
VERSION="$(fetch_requested_semver_level "${SEMVER_LEVEL}")"

if [[ "$(current_blob_newer_than "${VERSION}")" == "TRUE" ]];
then
  echo "MariaDB ${CURRENT} is the latest stable"
  ensure_blobstoreid_exists
else
  echo "Updating MariaDB from ${CURRENT} to ${VERSION}"
  replace_references_in_files "${VERSION}"
  replace_blobstore_version "${VERSION}"
fi
# TODO : PR workflow and blob removal if PR doesn't pass the tests
popd >/dev/null
