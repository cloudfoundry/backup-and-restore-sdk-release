#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

pushd "${SCRIPT_DIR}" &>/dev/null

echo "Checking latest patch release"
LATEST_PATCH_RELEASE="$(./list_new_patch_releases.sh  | tail -n 1)"
./bump_to_specific_version.sh "${LATEST_PATCH_RELEASE}"

popd &>/dev/null
