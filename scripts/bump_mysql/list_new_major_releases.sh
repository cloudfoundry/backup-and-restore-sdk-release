#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

pushd "${SCRIPT_DIR}" > /dev/null
CURRENT_VERSION="$1"
./list_new_stable_releases.sh "${CURRENT_VERSION}"
popd > /dev/null