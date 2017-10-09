#!/bin/bash

# Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.
#
# This program and the accompanying materials are made available under
# the terms of the under the Apache License, Version 2.0 (the "License”);
# you may not use this file except in compliance with the License.
#
# You may obtain a copy of the License at
# http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#
# See the License for the specific language governing permissions and
# limitations under the License.

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
export S3_BUCKET_NAME
export S3_REGION

cd src/github.com/cloudfoundry-incubator/backup-and-restore-sdk-release
glide install
ginkgo -r system_tests/blobstore
