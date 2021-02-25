#!/usr/bin/env bash

set -euo pipefail
shopt -s nocasematch

if [[ -z "${URL}" ]]
then
  echo "URL can't be empty."
  exit 1
fi

FLAGS=''

if [[ "${METHOD}" = "REBASE" ]]
then
  FLAGS+="--rebase "
elif [[ "${METHOD}" = "SQUASH" ]]
then
  FLAGS+="--squash "
elif [[ "${METHOD}" = "MERGE" ]]
then
  FLAGS+="--merge "
else
  echo "Only REBASE, SQUASH or MERGE are accepted values for METHOD. Received: ${METHOD}"
  exit 1
fi

if [[ "${DELETE}" = "TRUE" ]]
then
  FLAGS+="--delete-branch "
elif [[ "${DELETE}" = "FALSE" ]]
then
  : # Source branch won't be deleted
else
  echo "Only TRUE or FALSE are accepted values for DELETE. Received: ${DELETE}"
  exit 1
fi

gh auth login --with-token < <(echo $GITHUB_TOKEN)
gh pr merge "${URL}" ${FLAGS}