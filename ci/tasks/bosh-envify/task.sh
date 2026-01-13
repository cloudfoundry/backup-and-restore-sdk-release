#!/usr/bin/env bash

set -euo pipefail

[ -d toolsmiths-env ]
[ -d bosh-env ]

export ENVIRONMENT_LOCK_METADATA=toolsmiths-env/metadata

eval "$( smith bosh )"

readonly bosh_all_proxy_pattern='ssh\+socks5:\/\/(.*)@(([0-9]+\.){3}([0-9]+)):22\?private-key=(.*)'

# JUMPBOX_PRIVATE_KEY is present for cf-deployment pool envs
: "${JUMPBOX_PRIVATE_KEY:="$(echo "${BOSH_ALL_PROXY}" | sed -n -E "s/${bosh_all_proxy_pattern}/\5/p")"}"

if [ -f "$BOSH_CA_CERT" ]
then
  BOSH_CA_CERT="$(cat "$BOSH_CA_CERT")"
  # export BOSH_CA_CERT
fi

cat > bosh-env/alias-env.sh << EOF
export INSTANCE_JUMPBOX_PRIVATE="$(cat "${JUMPBOX_PRIVATE_KEY}")"
export INSTANCE_JUMPBOX_USER="$(echo "${BOSH_ALL_PROXY}" | sed -n -E "s/${bosh_all_proxy_pattern}/\1/p")"
export INSTANCE_JUMPBOX_EXTERNAL_IP="$(echo "${BOSH_ALL_PROXY}" | sed -n -E "s/${bosh_all_proxy_pattern}/\2/p")"

JUMPBOX_PRIVATE_KEY="\$(mktemp)"
chmod 0600 "\$JUMPBOX_PRIVATE_KEY"
echo "\$INSTANCE_JUMPBOX_PRIVATE" > "\$JUMPBOX_PRIVATE_KEY"

export BOSH_CLIENT="$BOSH_CLIENT"
export BOSH_CLIENT_SECRET="$BOSH_CLIENT_SECRET"
export BOSH_ENVIRONMENT="$BOSH_ENVIRONMENT"
export BOSH_CA_CERT="$BOSH_CA_CERT"
export BOSH_ALL_PROXY="ssh+socks5://\${INSTANCE_JUMPBOX_USER}@\${INSTANCE_JUMPBOX_EXTERNAL_IP}:22?private-key=\${JUMPBOX_PRIVATE_KEY}"
export BOSH_ENV_NAME="$(jq -r '.name' ${ENVIRONMENT_LOCK_METADATA})"

export CREDHUB_PROXY="\$BOSH_ALL_PROXY"
export CREDHUB_SERVER="\${BOSH_ENVIRONMENT}:8844"
export CREDHUB_CLIENT="\$BOSH_CLIENT"
export CREDHUB_SECRET="\$BOSH_CLIENT_SECRET"
export CREDHUB_CA_CERT="\$BOSH_CA_CERT"
EOF

# shellcheck disable=SC2001
cat > bosh-env/metadata.yml << EOF
INSTANCE_JUMPBOX_PRIVATE: |-
$(cat ${JUMPBOX_PRIVATE_KEY} | sed -E 's/(-+(BEGIN|END) RSA PRIVATE KEY-+) *| +/\1\n/g' |  sed 's/^/  /')
INSTANCE_JUMPBOX_USER: "$(echo "${BOSH_ALL_PROXY}" | sed -n -E "s/${bosh_all_proxy_pattern}/\1/p")"
INSTANCE_JUMPBOX_EXTERNAL_IP: "$(echo "${BOSH_ALL_PROXY}" | sed -n -E "s/${bosh_all_proxy_pattern}/\2/p")"
BOSH_CLIENT: "$BOSH_CLIENT"
BOSH_CLIENT_SECRET: "$BOSH_CLIENT_SECRET"
BOSH_ENVIRONMENT: "$BOSH_ENVIRONMENT"
BOSH_CA_CERT: |-
$(echo $BOSH_CA_CERT | sed -E 's/(-+(BEGIN|END) CERTIFICATE-+) *| +/\1\n/g' | sed 's/^/  /')
CREDHUB_PROXY: "$BOSH_ALL_PROXY"
CREDHUB_SERVER: "${BOSH_ENVIRONMENT}:8844"
CREDHUB_CLIENT: "$BOSH_CLIENT"
CREDHUB_SECRET: "$BOSH_CLIENT_SECRET"
CREDHUB_CA_CERT: |-
$(echo $BOSH_CA_CERT | sed -E 's/(-+(BEGIN|END) CERTIFICATE-+) *| +/\1\n/g' | sed 's/^/  /')
EOF
