#!/usr/bin/env bash

set -eu

pushd /backup-and-restore-sdk-release/src/database-backup-restore
  ginkgo -r -v -skipPackage system_tests
popd

