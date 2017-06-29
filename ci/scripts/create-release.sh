#!/usr/bin/env bash

set -e
set -x

export VERSION=$(cat version/number)

pushd backup-and-restore-sdk-release
  bosh-cli create-release --version $VERSION --tarball=../backup-and-restore-sdk-release-build/backup-and-restore-sdk-$VERSION.tgz --force
popd
