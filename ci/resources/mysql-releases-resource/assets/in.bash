#!/bin/bash

get_download_url() {
    local version="${1:?version not specified}"

    echo "https://downloads.mysql.com/archives/get/p/23/file/mysql-$version.tar.gz"
}

download_file() {
    local url="${1:?url not specified}"
    local destination_dir="${2:?destination_dir not specified}"

    cd $destination_dir
    wget $url
}

build_output() {
    local version="${1:?version not specified}"

    echo "[{\"version\": {\"ref\": \"$version\"}}]"
}