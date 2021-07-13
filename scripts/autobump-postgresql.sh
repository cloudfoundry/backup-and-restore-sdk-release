#!/usr/bin/env bash
set -euo pipefail

export VERSIONS_URL='https://ftp.postgresql.org/pub/source/'

HTML="$(curl -s -L "${VERSIONS_URL}")"
HREF="$(echo "${HTML}" | xmllint --html --xpath "//a/@href" - 2>/dev/null)"
ALL_VERSIONS="$(echo "${HREF}" | grep -Eo '[0-9]+(\.[0-9]+){1,2}')"

export BLOBS_PREFIX="postgres"
export ALL_VERSIONS
export DOWNLOAD_URL='https://ftp.postgresql.org/pub/source/v${VERSION}/postgresql-${VERSION}.tar.gz'
export DOWNLOADED_FILENAME='postgresql-${VERSION}.tar.gz'
