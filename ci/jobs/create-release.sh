#!/usr/bin/env bash

set -e
set -x

export VERSION=$(cat version/number)

pushd database-backup-and-restore-release
  bosh -n create release --version $VERSION --with-tarball --force
  mv dev_releases/database-backup-and-restore/* ../database-backup-and-restore-release-build/
popd
