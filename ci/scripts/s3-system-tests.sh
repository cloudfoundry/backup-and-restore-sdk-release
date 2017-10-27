#!/usr/bin/env bash

set -ex

eval "$(ssh-agent)"
./bosh-backup-and-restore-meta/unlock-ci.sh

chmod 400 bosh-backup-and-restore-meta/keys/github
ssh-add bosh-backup-and-restore-meta/keys/github
chmod 400 bosh-backup-and-restore-meta/genesis-bosh/bosh.pem

export GOPATH=$PWD
export PATH=$PATH:$GOPATH/bin
export BOSH_ENVIRONMENT="https://lite-bosh.backup-and-restore.cf-app.com"
export BOSH_CA_CERT=`pwd`/bosh-backup-and-restore-meta/certs/lite-bosh.backup-and-restore.cf-app.com.crt
export BOSH_GW_USER=vcap
export BOSH_GW_HOST=lite-bosh.backup-and-restore.cf-app.com
export BOSH_GW_PRIVATE_KEY=`pwd`/bosh-backup-and-restore-meta/genesis-bosh/bosh.pem
export AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY
export AWS_TEST_BUCKET_NAME
export AWS_TEST_BUCKET_REGION
export AWS_TEST_CLONE_BUCKET_NAME
export AWS_TEST_CLONE_BUCKET_REGION

cd src/github.com/cloudfoundry-incubator/backup-and-restore-sdk-release
glide install
ginkgo -r --focus=S3 system_tests