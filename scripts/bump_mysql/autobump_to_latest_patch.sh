#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

pushd "${SCRIPT_DIR}" &>/dev/null

echo "Checking latest patch release"
CURRENT_BLOBS_VERSION="$(./current_blobs_version.sh)"

for BLOB_VERSION in $CURRENT_BLOBS_VERSION;
do
    LATEST_PATCH_RELEASE="$(./list_new_patch_releases.sh "${BLOB_VERSION}"  | tail -n 1)"
    NEWVERSION="$(./download_specific_version.sh "${LATEST_PATCH_RELEASE}")"
    ./bump_to_specific_version.sh "${NEWVERSION}"
done

popd &>/dev/null
