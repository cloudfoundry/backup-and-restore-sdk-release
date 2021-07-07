#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

pushd "${SCRIPT_DIR}" &> /dev/null
REPO_ROOT="$(git rev-parse --show-toplevel)"
popd &> /dev/null


pushd "${REPO_ROOT}" &> /dev/null
CURRENT_BLOB_NAME="$(bosh blobs | grep 'mariadb' | cut -f1 | xargs)"
echo "${CURRENT_BLOB_NAME}"
popd &> /dev/null
