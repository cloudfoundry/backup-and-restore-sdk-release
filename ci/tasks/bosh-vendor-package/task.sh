#!/usr/bin/env bash

set -eux
set -o pipefail

pushd release
  set +x
  echo "---
blobstore:
  provider: s3 
  options:
    access_key_id: $AWS_ACCESS_KEY_ID
    secret_access_key: $AWS_SECRET_ACCESS_KEY
" > config/private.yml
  set -x

  bosh vendor-package "$PACKAGE_NAME" ../vendored-package-repo

  git add packages
  git add .final_builds

  git config --global user.name "$COMMIT_USER_NAME"
  git config --global user.email "$COMMIT_USER_EMAIL"

  if [ -n "$(git status --porcelain)" ]; then
    git commit -m "Update BOSH vendored package $PACKAGE_NAME"
  else
    echo "Vendored package $PACKAGE_NAME is already up-to-date";
  fi
popd

cp -R release/. updated-release
