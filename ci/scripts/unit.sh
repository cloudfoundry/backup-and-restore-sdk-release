#!/bin/bash

set -ex

eval "$(ssh-agent)"
./bosh-backup-and-restore-meta/unlock-ci.sh
chmod 400 bosh-backup-and-restore-meta/keys/github
ssh-add bosh-backup-and-restore-meta/keys/github
export GOPATH=$PWD
export PATH=$PATH:$GOPATH/bin


cd src/github.com/pivotal-cf/database-backup-and-restore-release
glide install
ginkgo unit-tests