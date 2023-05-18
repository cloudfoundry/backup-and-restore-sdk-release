#!/bin/bash
set -euo pipefail
set -x

eval "$(bbl print-env --metadata-file cf-deployment-env/metadata)"

function get_ip_port() {
    grep -o '[0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}:[0-9]\{1,\}' <<< "$1"
}

cat <<EOF > source-file/source-file.yml
---
jumpbox_username: jumpbox
jumpbox_url: $( get_ip_port "$BOSH_ALL_PROXY" )
jumpbox_ssh_key: |-
$(cat "$JUMPBOX_PRIVATE_KEY" | sed 's/^/  /')
target: ${BOSH_ENVIRONMENT}
client: ${BOSH_CLIENT}
client_secret: ${BOSH_CLIENT_SECRET}
ca_cert: |-
$(echo "$BOSH_CA_CERT" | sed 's/^/  /')
EOF

