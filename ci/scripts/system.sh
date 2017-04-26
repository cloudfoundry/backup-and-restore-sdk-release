#!/bin/bash

set -ex

eval "$(ssh-agent)"
./bosh-backup-and-restore-meta/unlock-ci.sh

chmod 400 bosh-backup-and-restore-meta/keys/github
ssh-add bosh-backup-and-restore-meta/keys/github
chmod 400 bosh-backup-and-restore-meta/genesis-bosh/bosh.pem


export GOPATH=$PWD
export PATH=$PATH:$GOPATH/bin
export BOSH_URL="https://genesis-bosh.backup-and-restore.cf-app.com:25555"
export BOSH_CERT_PATH=`pwd`/bosh-backup-and-restore-meta/certs/genesis-bosh.backup-and-restore.cf-app.com.crt
export BOSH_CLIENT
export BOSH_CLIENT_SECRET
export BOSH_GATEWAY_USER=vcap
export BOSH_GATEWAY_HOST=genesis-bosh.backup-and-restore.cf-app.com
export BOSH_GATEWAY_KEY=`pwd`/bosh-backup-and-restore-meta/genesis-bosh/bosh.pem
export POSTGRES_PASSWORD



cd src/github.com/pivotal-cf/database-backup-and-restore-release
glide install
ginkgo system-tests