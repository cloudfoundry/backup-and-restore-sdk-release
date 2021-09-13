#!/usr/bin/env bash
set -euo pipefail

export VERSIONS_URL='https://www.boost.org/users/history/'

HTML="$(curl -s -L "${VERSIONS_URL}")"
VALUES="$(echo "${HTML}" | xmllint --html --xpath "//h2[@class='news-title']/a[@href]/text()" - 2>/dev/null)"
ALL_VERSIONS="$(echo "${VALUES}" | grep -Eo '[0-9]+(\.[0-9]+){1,2}')"

export BLOBS_PREFIX="boost"
export ALL_VERSIONS
export DOWNLOADED_FILENAME='boost_${VERSION}.tar.gz'

function checksum_callback() {
    VERSION="${1}"
    DOWNLOADED_FILE="${2}"

    V_UNDERSCORES="$(echo "${VERSION}" | tr '.' '_')"
    CHECKSUM_HTML="$(curl -s -L "https://www.boost.org/users/history/version_${V_UNDERSCORES}.html")"
    EXPECTED_SHA256="$(echo "${CHECKSUM_HTML}" | xmllint --html --xpath "//td[a[contains(text(),'boost_${V_UNDERSCORES}.tar.gz')]]/following-sibling::td/text()" - 2>/dev/null)"
    echo "${EXPECTED_SHA256}  ${DOWNLOADED_FILE}" | sha256sum -c - || exit 1
}

function download_url_callback() {
    local VERSION="${1}"

    V_UNDERSCORES="$(echo "${VERSION}" | tr '.' '_')"
    DOWNLOAD_HTML="$(curl -s -L "https://www.boost.org/users/history/version_${V_UNDERSCORES}.html")"
    HREF="$(echo "${DOWNLOAD_HTML}" | xmllint --html --xpath "//td/a[contains(text(),'boost_${V_UNDERSCORES}.tar.gz')]/@href" - 2>/dev/null)"
    DOWNLOAD_URL="$(echo "${HREF}" | sed -r 's/href="(.+)"/\1/g' | sed -r 's/ //g' )"
    echo "${DOWNLOAD_URL}"
}

function extract_version_callback() {
    local BLOB_ID="${1}"

    local VERSION="$(echo "${BLOB_ID}" | grep -Eo '[0-9]+_[0-9]+_[0-9]+' | tr '_' '.')"
    echo "${VERSION}"
}

function new_version_callback() {
    local VERSION="${1}"

    local MAJOR="$(echo "${VERSION}" | grep -Eo '^[0-9]+')"
    local NEW_VERSION="$(echo "${ALL_VERSIONS}" |  sort -t "." -k1,1n -k2,2n -k3,3n | grep -Eo "^${MAJOR}.*" | tail -n 1)"

    # https://3.basecamp.com/4415260/buckets/20872488/question_answers/3963970683
    # MySQL 5.7 needs boost, and it needs it at 1.59.0 specifically
    echo "1.59.0"
    # echo "${NEW_VERSION}"
}

function new_blobid_callback() {
    local PRE_BLOBID="${1}"
    local PRE_VERSION="${2}"
    local NEW_VERSION="${3}"

    PREV_UNDERSCORES="$(echo "${PRE_VERSION}" | tr '.' '_')"
    NEWV_UNDERSCORES="$(echo "${NEW_VERSION}" | tr '.' '_')"
    NEW_BLOB_ID="$(echo "${PRE_BLOBID}" | sed "s/${PREV_UNDERSCORES}/${NEWV_UNDERSCORES}/")"
    echo "${NEW_BLOB_ID}"
}

function replace_references_callback() {
    local PRE_BLOB_ID="${1}"
    local NEW_BLOB_ID="${2}"
    local PRE_VERSION="${3}"
    local NEW_VERSION="${4}"
    local NEW_TARFILE="${5}"
    local BLOBS_PREFIX="${6}"

    local PREV_UNDERSCORES="$(echo "${PRE_VERSION}" | tr '.' '_')"
    local NEWV_UNDERSCORES="$(echo "${NEW_VERSION}" | tr '.' '_')"

    local FILES_WITH_REFS="$(grep -rnl '.' -e "${PREV_UNDERSCORES}" | grep "${BLOBS_PREFIX}")"

    for file in $FILES_WITH_REFS;
    do
    sed -i.bak "s#${PRE_BLOB_ID}#${NEW_BLOB_ID}#g" "$file"
    sed -i.bak "s#${PREV_UNDERSCORES}#${NEWV_UNDERSCORES}#g" "$file"
    rm "$file".bak
    done
}
