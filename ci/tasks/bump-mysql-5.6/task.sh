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

  current_blob_name=$(grep mysql-5.6 config/blobs.yml | tr -d :)

  bosh remove-blob $current_blob_name

  new_blob_name=$(ls ../mysql-5.6-release)

  echo $new_blob_name

  bosh add-blob ../mysql-5.6-release/$new_blob_name mysql/$new_blob_name
  bosh upload-blobs

  sed -i "s@$current_blob_name@mysql/$new_blob_name@" packages/database-backup-restorer-mysql-5.6/spec

  current_mysql_version=$(echo $current_blob_name | cut -c13-18)
  new_mysql_version=$(echo $new_blob_name | cut -c7-12)

  sed -i "s@$current_mysql_version@$new_mysql_version@" packages/database-backup-restorer-mysql-5.6/packaging

  git add .
  
  if [ -z "$VENDOR_UPDATES_BRANCH" ]
  then
        curr_branch=$(git rev-parse --abbrev-ref HEAD)
        echo "Pushing package updates to the same branch '${curr_branch}'"
  else
        git checkout -b "${VENDOR_UPDATES_BRANCH}"
        echo "Pushing package updates to the configured branch '${VENDOR_UPDATES_BRANCH}'"
  fi

  if [ -z "${COMMIT_USERNAME}" ] || [ -z "${COMMIT_USEREMAIL}" ]
  then
        echo "Unspecified user.name or user.email. Using defaults."
  else
        git config user.name "${COMMIT_USERNAME}"
        git config user.email "${COMMIT_USEREMAIL}"
  fi

  if git commit -m "Update blob mysql-5.6 from $current_mysql_version to $new_mysql_version"; then
    echo "Updated blob mysql-5.6"
  else
    echo "No change to blob mysql-5.6"
  fi
popd

cp -r release/. release-with-updated-vendored-package
