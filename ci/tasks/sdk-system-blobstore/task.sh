#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

if ls director-with-iam-profile/*.yml; then
  export BOSH_GW_PRIVATE_KEY=$( yq .jumpbox_ssh_key director-with-iam-profile/*.yml -r )
  export BOSH_GW_HOST_PORT=$( yq .jumpbox_url ${PWD}/director-with-iam-profile/*.yml -r )
  export BOSH_GW_HOST=$( echo ${BOSH_GW_HOST_PORT} | cut -f1 -d: )
  export BOSH_ENVIRONMENT=$( yq .target ${PWD}/director-with-iam-profile/*.yml -r )
  export BOSH_CLIENT_SECRET=$( yq .client_secret ${PWD}/director-with-iam-profile/*.yml -r )
  export BOSH_CLIENT=$( yq .client ${PWD}/director-with-iam-profile/*.yml -r )
  export BOSH_CA_CERT=$( yq .ca_cert ${PWD}/director-with-iam-profile/*.yml -r )
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
