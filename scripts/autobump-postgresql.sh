#!/usr/bin/env bash
set -euo pipefail

HTML="$(curl -s -L https://ftp.postgresql.org/pub/source/)"
HREF="$(echo "${HTML}" | xmllint --html --xpath "//a/@href" - 2>/dev/null)"
ALL_VERSIONS="$(echo "${HREF}" | grep -Eo '[0-9]+(\.[0-9]+){1,2}')"

export BLOBS_PREFIX="postgres"
export ALL_VERSIONS
export DOWNLOAD_URL='https://ftp.postgresql.org/pub/source/v${VERSION}/postgresql-${VERSION}.tar.gz'
export DOWNLOADED_FILENAME='postgresql-${VERSION}.tar.gz'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
pushd "${SCRIPT_DIR}" >/dev/null
source "./bump-generic/autobump-all-blobs-at-once.sh"
popd >/dev/null