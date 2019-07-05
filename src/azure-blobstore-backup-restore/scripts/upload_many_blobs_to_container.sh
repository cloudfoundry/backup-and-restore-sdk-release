#!/usr/bin/env bash

set -e
set -u

file="$(mktemp)"

function upload_blob() {
    local from="$1"
    local to="$2"

    for n in $(seq "$from" "$to"); do
        az storage blob upload \
        --container-name "$CONTAINER_NAME" \
        --name "test_$n" \
        --file "$file"
    done
}

#upload_blob 1 1000
#upload_blob 1001 2000&
#upload_blob 2001 3000&
#upload_blob 3001 4000&
#upload_blob 4001 5000
upload_blob 5001 6000&
upload_blob 6001 7000&
upload_blob 7001 8000&
upload_blob 8001 9000&
upload_blob 9001 10100