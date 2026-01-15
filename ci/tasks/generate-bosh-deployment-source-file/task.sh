#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

environment_metadata_json="cf-deployment-env/metadata"

bosh_all_proxy=$(jq -r '.bosh.bosh_all_proxy' ${environment_metadata_json})
bosh_ca_cert=$(jq -r '.bosh.bosh_ca_cert' ${environment_metadata_json})
bosh_client=$(jq -r '.bosh.bosh_client' ${environment_metadata_json})
bosh_client_secret=$(jq -r '.bosh.bosh_client_secret' ${environment_metadata_json})
bosh_environment=$(jq -r '.bosh.bosh_environment' ${environment_metadata_json})
jumpbox_private_key=$(jq -r '.bosh.jumpbox_private_key' ${environment_metadata_json})

function get_ip_port() {
    grep -o '[0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}:[0-9]\{1,\}' <<< "$1"
}

jumpbox_url=$( get_ip_port "${bosh_all_proxy}" )

# shellcheck disable=SC2001
cat <<EOF > source-file/source-file.yml
---
jumpbox_username: jumpbox
jumpbox_url: ${jumpbox_url}
jumpbox_ssh_key: |-
$(echo "${jumpbox_private_key}" | sed 's/^/  /')
target: ${bosh_environment}
client: ${bosh_client}
client_secret: ${bosh_client_secret}
ca_cert: |-
$(echo "${bosh_ca_cert}" | sed 's/^/  /')
EOF
