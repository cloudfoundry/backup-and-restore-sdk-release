#!/usr/bin/env bash
set -euo pipefail

export VERSIONS_URL='https://downloads.mysql.com/archives/community/'

HTML="$(curl -s -L "${VERSIONS_URL}")"
VALUES="$(echo "${HTML}" | xmllint --html --xpath "//select[@id='version']/option[@value=text()]/@value" - 2>/dev/null)"
ALL_VERSIONS="$(echo "${VALUES}" | grep -Eo '[0-9]+(\.[0-9]+){1,2}[a-zA-Z]?')"

export BLOBS_PREFIX="mysql"
export ALL_VERSIONS
# shellcheck disable=SC2016
export DOWNLOADED_FILENAME='$(basename "$(download_url_callback "${VERSION}")")'

function checksum_callback() {
    local version="${1}"
    local downloaded_file="${2}"
    local md5_url
    md5_url=$(download_url_callback "${version}").md5
    curl -Ls "${md5_url}" | md5sum -c -
}

function download_url_callback() {
    local version="${1}"
    local major_minor=${version%.*}
    local url
    case "${major_minor}" in
      "8.0") url=https://cdn.mysql.com/Downloads/MySQL-8.0/mysql-${version}-linux-glibc2.17-x86_64-minimal.tar.xz ;;
      "5.7") url=https://cdn.mysql.com/Downloads/MySQL-5.7/mysql-${version}.tar.gz ;;
      "5.6") url=https://cdn.mysql.com/Downloads/MySQL-5.6/mysql-${version}.tar.gz ;;
      *)
        >&2 echo "Unsupported MySQL version ${version}"
        return 1
        ;;
    esac
    echo "${url}"
}

function extract_version_callback() {
  local blob="${1:?}"

  echo "$blob " | grep -Eo '[0-9]+(\.[0-9]+){2}'
}

function new_version_callback() {
  echo "AUTO"
}
