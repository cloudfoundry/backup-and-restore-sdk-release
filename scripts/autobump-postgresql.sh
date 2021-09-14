#!/usr/bin/env bash
set -euo pipefail

export VERSIONS_URL='https://ftp.postgresql.org/pub/source/'

HTML="$(curl -s -L "${VERSIONS_URL}")"
HREF="$(echo "${HTML}" | xmllint --html --xpath "//a/@href" - 2>/dev/null)"
ALL_VERSIONS="$(echo "${HREF}" | grep -Eo '[0-9]+(\.[0-9]+){1,2}')"

export BLOBS_PREFIX="postgres"
export ALL_VERSIONS
export DOWNLOADED_FILENAME='postgresql-${VERSION}.tar.gz'

function checksum_callback() {
    VERSION="${1}"
    DOWNLOADED_FILE="${2}"

    EXPECTED_SHA256="$(curl -s -L "https://ftp.postgresql.org/pub/source/v${VERSION}/postgresql-${VERSION}.tar.gz.sha256" | cut -d ' ' -f1)"
    echo "${EXPECTED_SHA256}  ${DOWNLOADED_FILE}" | sha256sum -c - || exit 1
}

function download_url_callback() {
    local VERSION="${1}"
    echo "https://ftp.postgresql.org/pub/source/v${VERSION}/postgresql-${VERSION}.tar.gz"
}

function new_version_callback() {
  echo "AUTO"
}
