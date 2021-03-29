#!/usr/bin/env bash

set -eu

pushd backup-and-restore-sdk-release
  echo "---
blobstore:
  provider: s3
  options:
    access_key_id: $AWS_ACCESS_KEY_ID
    secret_access_key: $AWS_SECRET_ACCESS_KEY
" > config/private.yml

  current_blob_name=$(grep mysql-5.6 config/blobs.yml)
  current_blob_name=${blob_name%:}

  bosh remove-blob $current_blob_name

  new_blob_name=$(ls ../mysql-5.6-release)

  echo $new_blob_name

# bosh add-blob /Users/gbandres/Downloads/mysql-5.6.51.tar.gz mysql/mysql-5.6.51.tar.gz
# bosh upload-blobs

  # bosh vendor-package "${VENDORED_PACKAGE_NAME}" ../vendored-package-release

  # git add .
  
  # if [ -z "$VENDOR_UPDATES_BRANCH" ]
  # then
  #       curr_branch=$(git rev-parse --abbrev-ref HEAD)
  #       echo "Pushing package updates to the same branch '${curr_branch}'"
  # else
  #       git checkout -b "${VENDOR_UPDATES_BRANCH}"
  #       echo "Pushing package updates to the configured branch '${VENDOR_UPDATES_BRANCH}'"
  # fi

  # if [ -z "${COMMIT_USERNAME}" ] || [ -z "${COMMIT_USEREMAIL}" ]
  # then
  #       echo "Unspecified user.name or user.email. Using defaults."
  # else
  #       git config user.name "${COMMIT_USERNAME}"
  #       git config user.email "${COMMIT_USEREMAIL}"
  # fi

  # if git commit -m "Update package ${VENDORED_PACKAGE_NAME}"; then
  #   echo "Updated package ${VENDORED_PACKAGE_NAME}"
  # else
  #   echo "No change to vendored package ${VENDORED_PACKAGE_NAME}"
  # fi
popd

# cp -r release/. release-with-updated-vendored-package
