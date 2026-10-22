#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

if [[ -n "${BOSH_GW_HOST:-}" ]]; then
  ssh_key="$(mktemp)"
  echo -e "${BOSH_GW_PRIVATE_KEY}" > "${ssh_key}"
  chmod 0600 "${ssh_key}"
  export BOSH_GW_PRIVATE_KEY="${ssh_key}"
  export BOSH_ALL_PROXY="ssh+socks5://${BOSH_GW_USER}@${BOSH_GW_HOST}:22?private-key=${BOSH_GW_PRIVATE_KEY}"
fi

GCP_SERVICE_ACCOUNT_KEY_PATH="$(mktemp)"
echo "${GCP_SERVICE_ACCOUNT_KEY:-}" > "$GCP_SERVICE_ACCOUNT_KEY_PATH"
export GCP_SERVICE_ACCOUNT_KEY_PATH

pushd "backup-and-restore-sdk-release/src/system-tests/${TEST_SUITE_NAME}"
  if [[ -n "${FOCUS_SPEC:-}" ]]; then
    go run github.com/onsi/ginkgo/v2/ginkgo -mod vendor -focus "${FOCUS_SPEC}" -v -r --trace
  else
    go run github.com/onsi/ginkgo/v2/ginkgo -mod vendor -v -r --trace
  fi
popd
