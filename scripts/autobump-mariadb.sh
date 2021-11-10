#!/usr/bin/env bash
set -euo pipefail

set -x

export VERSIONS_URL='https://downloads.mariadb.org/mariadb/+releases/'

HTML="$(curl -s -L "${VERSIONS_URL}")"
VALUES="$(echo "${HTML}" | xmllint --html --xpath "//table/tbody/tr[td[3]='Stable']/td[1]/a/@href" - 2>/dev/null)"
ALL_VERSIONS="$(echo "${VALUES}" | grep -Eo '[0-9]+(\.[0-9]+){1,2}')"

export BLOBS_PREFIX="mariadb"
export ALL_VERSIONS
export DOWNLOADED_FILENAME='mariadb-${VERSION}.tar.gz'

function checksum_callback() {
    VERSION="${1}"
    DOWNLOADED_FILE="${2}"

    MAJOR_MINOR="$(echo "${VERSION}" | grep -Eo '[0-9]+\.[0-9]+')"
    CHECKSUM_JSON="$(curl -s -L "https://downloads.mariadb.org/rest-api/mariadb/${MAJOR_MINOR}/")"
    EXPECTED_SHA256="$(echo "${CHECKSUM_JSON}" | jq -r --arg v "${VERSION}" '.releases[$v].files[] | select(.os == "Source").checksum.sha256sum')"
    echo "${EXPECTED_SHA256}  ${DOWNLOADED_FILE}" | sha256sum -c - || exit 1
}

function download_url_callback() {
    local VERSION="${1}"
    echo "https://archive.mariadb.org//mariadb-${VERSION}/source/mariadb-${VERSION}.tar.gz"
}

function new_version_callback() {
  echo "AUTO"
}
