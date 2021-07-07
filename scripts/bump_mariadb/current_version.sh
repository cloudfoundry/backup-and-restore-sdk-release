#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

pushd "${SCRIPT_DIR}" &> /dev/null
CURRENT_BLOB_NAME="$(./current_blob_name.sh)"
CURRENT_BLOB_VERSION="$(echo "${CURRENT_BLOB_NAME}" | sed -n 's/^.*mariadb-\(.*\).tar.gz.*$/\1/p')"
echo "${CURRENT_BLOB_VERSION}"
popd &> /dev/null
