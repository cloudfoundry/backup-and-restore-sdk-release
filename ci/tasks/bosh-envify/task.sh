#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

export environment_metadata_json=toolsmiths-env/metadata

bosh_all_proxy=$(jq -r '.bosh.bosh_all_proxy' ${environment_metadata_json})
bosh_ca_cert=$(jq -r '.bosh.bosh_ca_cert' ${environment_metadata_json})
bosh_client=$(jq -r '.bosh.bosh_client' ${environment_metadata_json})
bosh_client_secret=$(jq -r '.bosh.bosh_client_secret' ${environment_metadata_json})
bosh_environment=$(jq -r '.bosh.bosh_environment' ${environment_metadata_json})
jumpbox_private_key=$(jq -r '.bosh.jumpbox_private_key' ${environment_metadata_json})
bosh_env_name=$(jq -r '.name' ${environment_metadata_json})

readonly bosh_all_proxy_pattern='ssh\+socks5:\/\/(.*)@(([0-9]+\.){3}([0-9]+)):22\?private-key=(.*)'

bosh_all_proxy_private_key_filename=$(echo "${bosh_all_proxy}" | sed -n -E "s/${bosh_all_proxy_pattern}/\5/p")
echo "${jumpbox_private_key}" > "${bosh_all_proxy_private_key_filename}"
instance_jumpbox_user=$(echo "${bosh_all_proxy}" | sed -n -E "s/${bosh_all_proxy_pattern}/\1/p")
instance_jumpbox_external_ip=$(echo "${bosh_all_proxy}" | sed -n -E "s/${bosh_all_proxy_pattern}/\2/p")

cat > bosh-env/alias-env.sh << EOF
export INSTANCE_JUMPBOX_PRIVATE="${jumpbox_private_key}"
export INSTANCE_JUMPBOX_USER="${instance_jumpbox_user}"
export INSTANCE_JUMPBOX_EXTERNAL_IP="${instance_jumpbox_external_ip}"

JUMPBOX_PRIVATE_KEY="\$(mktemp)"
chmod 0600 "\${JUMPBOX_PRIVATE_KEY}"
echo "\$INSTANCE_JUMPBOX_PRIVATE" > "\${JUMPBOX_PRIVATE_KEY}"

export BOSH_ALL_PROXY="ssh+socks5://\${INSTANCE_JUMPBOX_USER}@\${INSTANCE_JUMPBOX_EXTERNAL_IP}:22?private-key=\${JUMPBOX_PRIVATE_KEY}"
export BOSH_CA_CERT="${bosh_ca_cert}"
export BOSH_CLIENT="${bosh_client}"
export BOSH_CLIENT_SECRET="${bosh_client_secret}"
export BOSH_ENVIRONMENT="${bosh_environment}"
export BOSH_ENV_NAME="${bosh_env_name}"

export CREDHUB_PROXY="\$BOSH_ALL_PROXY"
export CREDHUB_SERVER="\${BOSH_ENVIRONMENT}:8844"
export CREDHUB_CLIENT="\$BOSH_CLIENT"
export CREDHUB_SECRET="\$BOSH_CLIENT_SECRET"
export CREDHUB_CA_CERT="\$BOSH_CA_CERT"
EOF

# shellcheck disable=SC2001
cat > bosh-env/metadata.yml << EOF
INSTANCE_JUMPBOX_PRIVATE: |-
$(echo "${jumpbox_private_key}" | sed -E 's/(-+(BEGIN|END) RSA PRIVATE KEY-+) *| +/\1\n/g' | sed 's/^/  /')
INSTANCE_JUMPBOX_USER: "${instance_jumpbox_user}"
INSTANCE_JUMPBOX_EXTERNAL_IP: "${instance_jumpbox_external_ip}"

BOSH_CLIENT: "${bosh_client}"
BOSH_CLIENT_SECRET: "${bosh_client_secret}"
BOSH_ENVIRONMENT: "${bosh_environment}"
BOSH_CA_CERT: |-
$(echo "${bosh_ca_cert}" | sed -E 's/(-+(BEGIN|END) CERTIFICATE-+) *| +/\1\n/g' | sed 's/^/  /')

CREDHUB_PROXY: "${bosh_all_proxy}"
CREDHUB_SERVER: "${bosh_environment}:8844"
CREDHUB_CLIENT: "${bosh_client}"
CREDHUB_SECRET: "${bosh_client_secret}"
CREDHUB_CA_CERT: |-
$(echo "${bosh_ca_cert}" | sed -E 's/(-+(BEGIN|END) CERTIFICATE-+) *| +/\1\n/g' | sed 's/^/  /')
EOF
