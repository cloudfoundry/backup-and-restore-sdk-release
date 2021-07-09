#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

pushd "${SCRIPT_DIR}" > /dev/null
CURRENT_VERSION="$1"
MINOR_PREFIX="$(echo "${CURRENT_VERSION}" | cut -d '.' -f1)"

STABLE_RELEASES="$(./list_new_stable_releases.sh "${CURRENT_VERSION}")"
MINOR_RELEASES="$(echo "${STABLE_RELEASES}" | grep "^${MINOR_PREFIX}" || true)"

echo "${MINOR_RELEASES}"
popd > /dev/null
