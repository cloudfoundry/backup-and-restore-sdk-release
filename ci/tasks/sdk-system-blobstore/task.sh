#!/usr/bin/env bash

set -eu

if ls director-with-iam-profile/*.yml; then
  BOSH_GW_PRIVATE_KEY=$(bosh int ${PWD}/director-with-iam-profile/*.yml --path /jumpbox_ssh_key )
  BOSH_GW_HOST_PORT=$(bosh int ${PWD}/director-with-iam-profile/*.yml --path /jumpbox_url )
  BOSH_ENVIRONMENT=$(bosh int ${PWD}/director-with-iam-profile/*.yml --path /target )
  BOSH_CLIENT_SECRET=$(bosh int ${PWD}/director-with-iam-profile/*.yml --path /client_secret )
  BOSH_CLIENT=$(bosh int ${PWD}/director-with-iam-profile/*.yml --path /client )
  BOSH_CA_CERT=$(bosh int ${PWD}/director-with-iam-profile/*.yml --path /ca_cert )
fi

echo -e "${BOSH_GW_PRIVATE_KEY}" > "${PWD}/ssh.key"
chmod 0600 "${PWD}/ssh.key"
export BOSH_GW_PRIVATE_KEY="${PWD}/ssh.key"
export BOSH_ALL_PROXY="ssh+socks5://${BOSH_GW_USER}@${BOSH_GW_HOST}:22?private-key=${BOSH_GW_PRIVATE_KEY}"

GCP_SERVICE_ACCOUNT_KEY_PATH="$(mktemp)"
echo "$GCP_SERVICE_ACCOUNT_KEY" > "$GCP_SERVICE_ACCOUNT_KEY_PATH"
export GCP_SERVICE_ACCOUNT_KEY_PATH

pushd "backup-and-restore-sdk-release/src/system-tests/${TEST_SUITE_NAME}"
  if [[ -n "${FOCUS_SPEC}" ]]; then
    go run github.com/onsi/ginkgo/v2/ginkgo -mod vendor -focus "${FOCUS_SPEC}" -v -r --trace
  else
    go run github.com/onsi/ginkgo/v2/ginkgo -mod vendor -v -r --trace
  fi
popd
