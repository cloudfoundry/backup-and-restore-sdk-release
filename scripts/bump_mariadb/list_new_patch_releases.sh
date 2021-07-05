#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

pushd "${SCRIPT_DIR}" > /dev/null
CURRENT_VERSION="$(./current_version.sh)"
PATCH_PREFIX="$(echo "${CURRENT_VERSION}" | cut -d '.' -f1,2)"

STABLE_RELEASES="$(./list_new_stable_releases.sh)"
PATCH_RELEASES="$(echo "${STABLE_RELEASES}" | grep "^${PATCH_PREFIX}" || true)"

echo "${PATCH_RELEASES}"
popd > /dev/null
