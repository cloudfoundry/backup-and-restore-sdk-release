#!/usr/bin/env bash

set -e

GINKGO_EXTRA_FLAGS='-p --skipPackage s3bucket'

pushd "/backup-and-restore-sdk-release/src/s3-blobstore-backup-restore"
  ginkgo_cmd="ginkgo -mod vendor -r -keepGoing"

  if [[ -n "$GINKGO_EXTRA_FLAGS" ]]; then
    ginkgo_cmd="$ginkgo_cmd $GINKGO_EXTRA_FLAGS"
  fi

  set -x
  $ginkgo_cmd
  set +x
popd
