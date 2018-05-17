#!/usr/bin/env bash

set -ex

eval "$(ssh-agent)"
./bosh-backup-and-restore-meta/unlock-ci.sh

chmod 400 bosh-backup-and-restore-meta/keys/github
ssh-add bosh-backup-and-restore-meta/keys/github
chmod 400 bosh-backup-and-restore-meta/genesis-bosh/bosh.pem

export GOPATH="${PWD}/backup-and-restore-sdk-release"
export PATH="$PATH:$GOPATH/bin"
export BOSH_ENVIRONMENT="${BOSH_ENVIRONMENT:="https://lite-bosh.backup-and-restore.cf-app.com"}"
export BOSH_CA_CERT="${BOSH_CA_CERT:="${PWD}/bosh-backup-and-restore-meta/certs/lite-bosh.backup-and-restore.cf-app.com.crt"}"
export BOSH_GW_USER="${BOSH_GW_USER:="vcap"}"
export BOSH_GW_HOST="${BOSH_GW_HOST:="lite-bosh.backup-and-restore.cf-app.com"}"

if [[ -z "${BOSH_GW_PRIVATE_KEY}" ]]; then
  export BOSH_GW_PRIVATE_KEY="${PWD}/bosh-backup-and-restore-meta/genesis-bosh/bosh.pem"
else
  echo -e "${BOSH_GW_PRIVATE_KEY}" > "${PWD}/ssh.key"
  chmod 0600 "${PWD}/ssh.key"
  export BOSH_GW_PRIVATE_KEY="${PWD}/ssh.key"
fi

cd backup-and-restore-sdk-release/src/github.com/cloudfoundry-incubator/s3-blobstore-backup-restore

if [[ ! -z "${FOCUS_SPEC}" ]]; then
  ginkgo -focus "${FOCUS_SPEC}" -v -r system_tests -trace
else
  ginkgo -v -r system_tests -trace
fi

