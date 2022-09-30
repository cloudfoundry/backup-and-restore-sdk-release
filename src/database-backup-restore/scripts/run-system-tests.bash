#!/usr/bin/env bash

set -euo pipefail

SRC_DIR="$(cd "$( dirname "$0" )/.." && pwd)"

set -x
pushd "$SRC_DIR"
  ginkgo -mod vendor -r -v "system_tests/mysql" -trace
popd
