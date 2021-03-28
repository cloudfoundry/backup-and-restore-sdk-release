#!/bin/bash

get_latest_release() {
    local product_title="${1:?product_title not specified}"
    local xml_input=$(timeout 0.5s cat)

    local version_number=$(echo $xml_input | xq --raw-output -c '.rss.channel.item[] | select(.title | contains("MySQL Community Server 5.6")) | .description | split("(") | .[1] | split(" ") | .[0]')

    echo $version_number
}

build_output() {
    local version="${1:?version not specified}"

    echo "[{\"ref\": \"$version\"}]"
}