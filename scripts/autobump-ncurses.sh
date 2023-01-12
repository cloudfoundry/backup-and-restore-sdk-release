#!/usr/bin/env bash
set -euo pipefail

SCRIPT_RELATIVE_DIR="$(dirname "$0")"

export BLOBS_PREFIX="ncurses"

export VERSIONS_URL='ftp://ftp.gnu.org/pub/gnu/ncurses/'

DIRECTORY_LISTING="$(curl -s -L "${VERSIONS_URL}")"
ALL_VERSIONS="$(echo ${DIRECTORY_LISTING} | grep ncurses | grep -Eo '[0-9]+(\.[0-9]+){1,2}[a-zA-Z]?' | sort | uniq)"
export ALL_VERSIONS

export DOWNLOADED_FILENAME='ncurses-${VERSION}.tar.gz' # The docs say this is irrelevant but currently mandatory

function checksum_callback() {
    local version="${1}"
    local downloaded_file="${2}"
    local gpg_sig_url
    gpg_sig_url=$(download_url_callback "${version}").sig

    # GNU GPG Keyring originally downloaded from ftp://ftp.gnu.org/gnu/gnu-keyring.gpg
    gpg --quiet --import "$SCRIPT_RELATIVE_DIR/gnu-keyring.gpg"

    curl -Ls "${gpg_sig_url}" -o "${downloaded_file}.sig" \
         | gpg --verify - "${downloaded_file}"
}

function download_url_callback() {
    local version="${1}"
    echo "ftp://ftp.gnu.org/pub/gnu/ncurses/ncurses-${version}.tar.gz"
}

function new_version_callback() {
  echo "AUTO"
}
