#!/usr/bin/env bash
set -euo pipefail

pushd backup-and-restore-sdk-release > /dev/null
  echo "---
blobstore:
  provider: s3
  options:
    access_key_id: ${AWS_ACCESS_KEY_ID:-}
    secret_access_key: ${AWS_SECRET_ACCESS_KEY:-}
" > config/private.yml

  current_blob_name="$(scripts/bump_mariadb/current_blob_name.sh)"
  scripts/bump_mariadb/bump_to_specific_version.sh ../mariadb-release/*
  new_blob_name="$(scripts/bump_mariadb/current_blob_name.sh)"

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

  if git commit -m "Update blob from ${current_blob_name} to ${new_blob_name}"; then
    echo "Updated blob from ${current_blob_name} to ${new_blob_name}"
  else
    echo "No change to ${current_blob_name}"
  fi
popd > /dev/null

cp -r backup-and-restore-sdk-release/. updated-backup-and-restore-sdk-release
