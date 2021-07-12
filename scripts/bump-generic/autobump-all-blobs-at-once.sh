#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
SDK_ROOT="$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel)"

: "${BLOBS_PREFIX}"
: "${ALL_VERSIONS}"
: "${DOWNLOAD_URL}"
: "${DOWNLOADED_FILENAME}"

function current_blobs_names() {
    BLOBS_PREFIX="$1"
    CURRENT_BLOBS_NAMES="$(bosh blobs | grep "^${BLOBS_PREFIX}/" | cut -f1)"
    for BLOB_NAME in  $CURRENT_BLOBS_NAMES;
    do
        echo "${BLOB_NAME}" | xargs
    done
}

function download_version() {
    VERSION="$1"
    eval DOWNLOAD_URL="$2"
    eval DOWNLOAD_DESTINATION="$3"

    wget -q -O "${DOWNLOAD_DESTINATION}" "${DOWNLOAD_URL}"
    # TODO: Implement checksum
    realpath "${DOWNLOAD_DESTINATION}"
}


pushd "${SDK_ROOT}" >/dev/null
CURRENT_BLOBS_NAME="$(current_blobs_names "${BLOBS_PREFIX}")"
popd > /dev/null

pushd "${SCRIPT_DIR}" &>/dev/null
for BLOB_ID in $CURRENT_BLOBS_NAME;
do
    CUR_VERSION="$(echo "${BLOB_ID}" | grep -Eo '[0-9]+(\.[0-9]+)*')"
    NEW_VERSION="$(./autobump-pick-candidate.sh "${CUR_VERSION}" "${ALL_VERSIONS}")"
    NEW_TARFILE="$(download_version "${NEW_VERSION}" "${DOWNLOAD_URL}" "${SDK_ROOT}/${DOWNLOADED_FILENAME}")"
    NEW_BLOB_ID="$(echo "${BLOB_ID}" | grep "${CUR_VERSION}" | sed "s/${CUR_VERSION}/${NEW_VERSION}/")"

    ./replace-blob-with.sh "${BLOB_ID}" "${NEW_BLOB_ID}" "${CUR_VERSION}" "${NEW_VERSION}" "${NEW_TARFILE}"
    rm -f "${NEW_TARFILE}"
done
popd &>/dev/null
