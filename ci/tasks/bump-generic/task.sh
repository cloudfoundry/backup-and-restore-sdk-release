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

  blobs_before_autobump="$(bosh blobs | cut -f1 | grep -Eo '[0-9]+(\.[0-9]+)*')"
  ./"${AUTOBUMP_SCRIPT}"
  blobs_after_autobump="$(bosh blobs | cut -f1 | grep -Eo '[0-9]+(\.[0-9]+)*')"

  updated_blobs_old_version="$(diff <(echo "$blobs_before_autobump") <(echo "$blobs_after_autobump") | { grep "<" || true; } | sed 's/</ /g' | tr '\n' ' ' | xargs || true)"
  updated_blobs_new_version="$(diff <(echo "$blobs_before_autobump") <(echo "$blobs_after_autobump") | { grep ">" || true; } | sed 's/>/ /g' | tr '\n' ' ' | xargs || true)"

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

  if git commit -m "Update blobs from ${updated_blobs_old_version} to ${updated_blobs_new_version}"; then
    echo "Updated blobs from ${updated_blobs_old_version} to ${updated_blobs_new_version}"
  else
    echo "No change to blobs ${blobs_before_autobump}"
  fi
  exit 1
popd > /dev/null

cp -r backup-and-restore-sdk-release/. updated-backup-and-restore-sdk-release
