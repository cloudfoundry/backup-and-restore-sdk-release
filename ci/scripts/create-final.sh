#!/usr/bin/env bash

set -ex

VERSION=$(cat version/number)

pushd backup-and-restore-sdk-release
  echo "---
blobstore:
provider: s3
options:
  access_key_id: $AWS_ACCESS_KEY_ID
  secret_access_key: $AWS_SECRET_ACCESS_KEY
" > config/private.yml

  bosh -n finalize release ../release-tarball/backup-and-restore-sdk-*.tgz --version "$VERSION"

  mv releases/backup-and-restore-sdk/backup-and-restore-sdk-"${VERSION}".tgz \
    ../backup-and-restore-sdk-final-release-tarball

  git add ./releases
  git add ./.final_builds

  git config --global user.name "Backup & Restore Concourse"
  git config --global user.email "cf-lazarus@pivotal.io"

  git commit -m "Add final release ${VERSION} [ci skip]"
popd

cp -R backup-and-restore-sdk-release/. backup-and-restore-sdk-final-release
