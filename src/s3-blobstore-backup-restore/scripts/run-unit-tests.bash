#!/usr/bin/env bash

set -euo pipefail

SRC_DIR="$(cd "$( dirname "$0" )/.." && pwd)"

set -x
pushd "$SRC_DIR"
  go run github.com/onsi/ginkgo/v2/ginkgo -mod vendor -r --keep-going -p --skip-package s3bucket
popd
