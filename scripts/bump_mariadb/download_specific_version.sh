#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
SDK_ROOT="$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel)"

function download_version() {
  VERSION="$1"
  OUTFILE="mariadb-${VERSION}.tar.gz"
  wget -q -O "${OUTFILE}" "https://downloads.mariadb.org/interstitial/mariadb-${VERSION}/source/mariadb-${VERSION}.tar.gz"
  echo "${OUTFILE}"
}

VERSION="${1}"

pushd "${SDK_ROOT}" >/dev/null
download_version "${VERSION}"
popd > /dev/null
