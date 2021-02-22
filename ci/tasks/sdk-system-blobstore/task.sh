#!/usr/bin/env bash

set -eu

echo -e "${BOSH_GW_PRIVATE_KEY}" > "${PWD}/ssh.key"
chmod 0600 "${PWD}/ssh.key"
export BOSH_GW_PRIVATE_KEY="${PWD}/ssh.key"
export BOSH_ALL_PROXY="ssh+socks5://${BOSH_GW_USER}@${BOSH_GW_HOST}:22?private-key=${BOSH_GW_PRIVATE_KEY}"

GCP_SERVICE_ACCOUNT_KEY_PATH="$(mktemp)"
echo "$GCP_SERVICE_ACCOUNT_KEY" > "$GCP_SERVICE_ACCOUNT_KEY_PATH"
export GCP_SERVICE_ACCOUNT_KEY_PATH

pushd "backup-and-restore-sdk-release/src/system-tests/${TEST_SUITE_NAME}"
  if [[ -n "${FOCUS_SPEC}" ]]; then
    ginkgo -mod vendor -focus "${FOCUS_SPEC}" -v -r -trace
  else
    ginkgo -mod vendor -v -r -trace
  fi
popd
