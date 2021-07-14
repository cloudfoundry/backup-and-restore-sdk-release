#!/usr/bin/env bash
set -euo pipefail

export VERSIONS_URL='https://downloads.mysql.com/archives/community/'

HTML="$(curl -s -L "${VERSIONS_URL}")"
VALUES="$(echo "${HTML}" | xmllint --html --xpath "//select[@id='version']/option[@value=text()]/@value" - 2>/dev/null)"
ALL_VERSIONS="$(echo "${VALUES}" | grep -Eo '[0-9]+(\.[0-9]+){1,2}[a-zA-Z]?')"

export BLOBS_PREFIX="mysql"
export ALL_VERSIONS
export DOWNLOADED_FILENAME='mysql-${VERSION}.tar.gz'

function checksum_callback() {
    VERSION="${1}"
    DOWNLOADED_FILE="${2}"

    CHECKSUM_HTML="$(curl -s -L "https://downloads.mysql.com/archives/community/?version=${VERSION}&os=src&osva=Generic+Linux+%28Architecture+Independent%29#downloads")"
    EXPECTED_MD5="$(echo "${CHECKSUM_HTML}" | xmllint --html --xpath "//td[a/@href='/archives/gpg/?file=mysql-${VERSION}.tar.gz&p=23']/code/text()" - 2>/dev/null)"
    echo "${EXPECTED_MD5}  ${DOWNLOADED_FILE}" | md5sum -c - || exit 1
}

function download_url_callback() {
    local VERSION="${1}"
    echo "https://downloads.mysql.com/archives/get/p/23/file/mysql-${VERSION}.tar.gz"
}
