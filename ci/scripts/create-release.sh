#!/usr/bin/env bash

set -e
set -x

export VERSION=$(cat version/number)

pushd database-backup-and-restore-release
  bosh-cli create-release --version $VERSION --tarball=../database-backup-and-restore-release-build/database-backup-and-restore-release-$VERSION.tgz --force
popd
