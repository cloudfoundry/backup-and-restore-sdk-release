#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
SDK_ROOT="$(cd "${SCRIPT_DIR}/.." &>/dev/null && pwd)"

pushd "${SDK_ROOT}" >/dev/null

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

function last_stable_release() {
  stable_releases | sort -t "." -k1,1n -k2,2n -k3,3n | tail -n 1
}

function download_version() {
  VERSION="$1"
  OUTFILE="mariadb-${VERSION}.tar.gz"
  wget -q -O "${OUTFILE}" "https://downloads.mariadb.org/interstitial/mariadb-${VERSION}/source/mariadb-${VERSION}.tar.gz"
  echo "${OUTFILE}"
}

function update_blobstore() {
  NEW_BLOB_VERSION="$(last_stable_release)"
  NEW_BLOB_FILE=$(download_version "${NEW_BLOB_VERSION}")

  bosh remove-blob "--dir=${SDK_ROOT}" "$(current_blob_name)"
  bosh add-blob "--dir=${SDK_ROOT}" "${SDK_ROOT}/${NEW_BLOB_FILE}" "mariadb/mariadb-${NEW_BLOB_VERSION}.tar.gz"
  bosh upload-blobs "--dir=${SDK_ROOT}"
  rm "${NEW_BLOB_FILE}"
}

# TODO : Check if last-stable-release is newer than existing blob
# TODO : Only update blobstore if the last-stable-release is newer
# TODO : Allow bumping only PATCH/MINORs
update_blobstore
# TODO : Update hardcoded references to mariadb version
# TODO : PR workflow and blob removal if PR doesn't pass the tests
popd >/dev/null
