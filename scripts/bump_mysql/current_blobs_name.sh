#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

pushd "${SCRIPT_DIR}" &> /dev/null
REPO_ROOT="$(git rev-parse --show-toplevel)"
popd &> /dev/null


pushd "${REPO_ROOT}" &> /dev/null
CURRENT_BLOBS_NAME="$(bosh blobs | grep 'mysql' | cut -f1)"
for BLOB_NAME in  $CURRENT_BLOBS_NAME;
do
    echo "${BLOB_NAME}" | xargs
done
popd &> /dev/null
