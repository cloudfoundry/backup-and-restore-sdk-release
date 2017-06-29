#!/bin/bash

set -ex

eval "$(ssh-agent)"
./bosh-backup-and-restore-meta/unlock-ci.sh
chmod 400 bosh-backup-and-restore-meta/keys/github
ssh-add bosh-backup-and-restore-meta/keys/github

export GOPATH=`pwd`/backup-and-restore-sdk-release:"$GOPATH"

cd backup-and-restore-sdk-release/src/github.com/pivotal-cf/database-backup-and-restore
glide install --strip-vendor
ginkgo -r -v