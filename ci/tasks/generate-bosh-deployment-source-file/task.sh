#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

download_url="$(wget --header="Authorization: Bearer ${GITHUB_TOKEN}" -q -O - https://api.github.com/repos/cloudfoundry/bosh-bootloader/releases/tags/v${BBL_VERSION} | jq -r '.assets[] | select(.name | endswith("linux_amd64")) | .browser_download_url')"
wget --header="Authorization: Bearer ${GITHUB_TOKEN}" -q -O /usr/local/bin/bbl "$download_url"
chmod +x /usr/local/bin/bbl
bbl -v

eval "$(bbl print-env --metadata-file cf-deployment-env/metadata)"

function get_ip_port() {
    grep -o '[0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}:[0-9]\{1,\}' <<< "$1"
}

# shellcheck disable=SC2001
cat <<EOF > source-file/source-file.yml
---
jumpbox_username: jumpbox
jumpbox_url: $( get_ip_port "${BOSH_ALL_PROXY}" )
jumpbox_ssh_key: |-
$(cat "${JUMPBOX_PRIVATE_KEY}" | sed 's/^/  /')
target: ${BOSH_ENVIRONMENT}
client: ${BOSH_CLIENT}
client_secret: ${BOSH_CLIENT_SECRET}
ca_cert: |-
$(echo "${BOSH_CA_CERT}" | sed 's/^/  /')
EOF
