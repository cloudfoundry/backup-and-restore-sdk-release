#!/usr/bin/env bash

set -ex

eval "$(ssh-agent)"
./bosh-backup-and-restore-meta/unlock-ci.sh

chmod 400 bosh-backup-and-restore-meta/keys/github
ssh-add bosh-backup-and-restore-meta/keys/github
chmod 400 bosh-backup-and-restore-meta/genesis-bosh/bosh.pem

export GOPATH=$PWD/backup-and-restore-sdk-release
export PATH=$PATH:$GOPATH/bin
export BOSH_ENVIRONMENT="https://lite-bosh.backup-and-restore.cf-app.com"
export BOSH_CA_CERT=`pwd`/bosh-backup-and-restore-meta/certs/lite-bosh.backup-and-restore.cf-app.com.crt
export BOSH_GW_USER=vcap
export BOSH_GW_HOST=lite-bosh.backup-and-restore.cf-app.com
export BOSH_GW_PRIVATE_KEY=`pwd`/bosh-backup-and-restore-meta/genesis-bosh/bosh.pem

cd backup-and-restore-sdk-release/src/github.com/cloudfoundry-incubator/s3-blobstore-backup-restore
ginkgo -v -r system_tests -trace
