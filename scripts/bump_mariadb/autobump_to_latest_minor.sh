#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

pushd "${SCRIPT_DIR}" &>/dev/null

echo "Checking latest minor release"
LATEST_MINOR_RELEASE="$(./list_new_minor_releases.sh  | tail -n 1)"
NEWVERSION="$(./download_specific_version.sh "${LATEST_MINOR_RELEASE}")"
./bump_to_specific_version.sh "${NEWVERSION}"

popd &>/dev/null
