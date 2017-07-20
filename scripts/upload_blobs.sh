#!/bin/bash -eu

cd $(dirname $0)/..

lpass show private.yml --notes > config/private.yml

bosh-cli upload-blobs
