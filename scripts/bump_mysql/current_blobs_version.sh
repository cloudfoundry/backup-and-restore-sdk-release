#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

pushd "${SCRIPT_DIR}" &> /dev/null
CURRENT_BLOBS_NAME="$(./current_blobs_name.sh)"
CURRENT_BLOBS_VERSION="$(echo "${CURRENT_BLOBS_NAME}" | sed -n 's/^.*mysql-\(.*\).tar.gz.*$/\1/p')"
echo "${CURRENT_BLOBS_VERSION}"
popd &> /dev/null
