#!/usr/bin/env bash

set -euo pipefail

SRC_DIR="$(cd "$( dirname "$0" )/.." && pwd)"

set -x
pushd "$SRC_DIR"
  go run github.com/onsi/ginkgo/v2/ginkgo -r -v --skip-package system_tests
popd
