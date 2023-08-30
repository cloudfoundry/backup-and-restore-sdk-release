#!/usr/bin/env bash
set -uo pipefail

set -x

export VERSIONS_URL='https://ftp.openssl.org/source'

CURRENT_VERSIONS=$(curl -s -L ${VERSIONS_URL} | xmllint --html --xpath "//table/tr/td[2]/a/@href" - 2>/dev/null | grep -Eo '[0-9]+(\.[0-9]+){1,2}[a-z]' | sort -u)
OLD_VERSIONS=$(curl -s -L ${VERSIONS_URL}/old/1.1.1/ | xmllint --html --xpath "//table/tr/td[2]/a/@href" - 2>/dev/null | grep -Eo '[0-9]+(\.[0-9]+){1,2}[a-z]' | sort -u)
ALL_VERSIONS="$(echo -e "${OLD_VERSIONS}\n${CURRENT_VERSIONS}")"

export DOWNLOADED_FILENAME='openssl-${VERSION}.tar.gz'
export BLOBS_PREFIX=openssl

function checksum_callback() {
    VERSION="${1}"
    DOWNLOADED_FILE="${2}"

    #check if old or new version because the old versions have a different path than the new versions
    # the output for curl needs to be redirected to /dev/null, because we call the callback like this:  RETURN=$(function checksum_callback 1.1.1)
    if curl -s ${VERSIONS_URL} |grep "${VERSION}" &2> /dev/null; then
      DOWNLOAD_URL="${VERSIONS_URL}/openssl-${VERSION}.tar.gz"
      SHA_URL="${DOWNLOAD_URL}.sha256"
    else
      DOWNLOAD_URL="${VERSIONS_URL}/old/1.1.1/openssl-${VERSION}.tar.gz"
      SHA_URL="${DOWNLOAD_URL}.sha256"
    fi
    MAJOR_MINOR="$(echo "${VERSION}" | grep -Eo '[0-9]+\.[0-9]+')"
    echo "$(curl -s ${SHA_URL}) ${DOWNLOADED_FILE}" | sha256sum -c - || exit 1
}

function download_url_callback() {
    local VERSION="${1}"

    if curl -s -L ${VERSIONS_URL} | grep "${VERSION}" &2> /dev/null; then
      DOWNLOAD_URL="${VERSIONS_URL}/openssl-${VERSION}.tar.gz"
    else
      DOWNLOAD_URL="${VERSIONS_URL}/old/1.1.1/openssl-${VERSION}.tar.gz"
    fi

    echo "${DOWNLOAD_URL}"
}

function new_blobid_callback() {
    local PRE_BLOBID="${1}"
    local NEW_VERSION="${3}"

    echo "openssl/openssl-${NEW_VERSION}.tar.gz"
}

function extract_version_callback() {
  echo $(grep -Eo '[0-9]+(\.[0-9]+){1,2}[a-z]' <<<$1)
}
function new_version_callback() {
  echo "AUTO"
}

