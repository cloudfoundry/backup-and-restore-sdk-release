#!/usr/bin/env bash

set -euxo pipefail

BOSH_CREDS_SCRIPT="$1"
echo "Waiting for the BOSH Director to be ready"
until [ -f "${BOSH_CREDS_SCRIPT}" ]
do
  sleep 60
done
source "${BOSH_CREDS_SCRIPT}"

STEMCELL_URL="$(curl -L https://bosh.io/stemcells | grep -io "https:\/\/.*warden-boshlite-${STEMCELL_NAME}-go_agent.tgz")"

bosh -n --tty upload-stemcell "${STEMCELL_URL}"
pushd /backup-and-restore-sdk-release
  bosh -n --tty upload-release
popd

cat > manifest.yml <<EOF
---
name: compilation
releases:
- name: backup-and-restore-sdk
  version: 'latest'
stemcells:
- alias: default
  os: "$STEMCELL_NAME"
  version: 'latest'
update:
  canaries: 1
  max_in_flight: 1
  canary_watch_time: 1000 - 90000
  update_watch_time: 1000 - 90000
instance_groups: []
EOF

bosh -n --tty -d compilation deploy manifest.yml
export STEMCELL_VERSION="$(bosh -n stemcells --json | jq -r --arg stemcell "${STEMCELL_NAME}" '.Tables[0].Rows[] | select(.os==$stemcell).version' | sed 's/*//g')"
export RELEASE_VERSION="$(bosh -n releases --json | jq -r --arg release "backup-and-restore-sdk" '.Tables[0].Rows[] | select(.name==$release).version' | sed 's/*//g')"

bosh -n --tty -d compilation export-release backup-and-restore-sdk/$RELEASE_VERSION $STEMCELL_NAME/$STEMCELL_VERSION
