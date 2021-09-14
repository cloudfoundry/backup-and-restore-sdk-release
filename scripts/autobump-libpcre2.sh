#!/usr/bin/env bash
set -euo pipefail

export VERSIONS_URL='https://ftp.pcre.org/pub/pcre/'

HTML="$(curl -s -L "${VERSIONS_URL}")"
VALUES="$(echo "${HTML}" | grep -Eo 'pcre2-[0-9]+\.[0-9]+\.tar.gz' | uniq)"
ALL_VERSIONS="$(echo "${VALUES}" | grep -Eo '[0-9]+\.[0-9]+')"

export BLOBS_PREFIX="libpcre2"
export ALL_VERSIONS
export DOWNLOADED_FILENAME='pcre2-${VERSION}.tar.gz'

function download_url_callback() {
    local VERSION="${1}"
    echo "https://ftp.pcre.org/pub/pcre/pcre2-${VERSION}.tar.gz"
}

function extract_version_callback() {
    local BLOB_ID="${1}"
    echo "${BLOB_ID}" | grep -Eo '[0-9]+\.[0-9]+'
}

function new_version_callback() {
  echo "AUTO"
}
