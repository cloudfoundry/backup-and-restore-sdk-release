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
chmod 400 bosh-backup-and-restore-meta/garden-bosh-uaa/bosh.pem

export GOPATH=$PWD
export PATH=$PATH:$GOPATH/bin
export BOSH_URL="https://lite-bosh.backup-and-restore.cf-app.com"
export BOSH_CERT_PATH=`pwd`/bosh-backup-and-restore-meta/certs/lite-bosh.backup-and-restore.cf-app.com.crt
export BOSH_CLIENT
export BOSH_CLIENT_SECRET
export BOSH_GATEWAY_USER=vcap
export BOSH_GATEWAY_HOST=lite-bosh.backup-and-restore.cf-app.com
export BOSH_GATEWAY_KEY=`pwd`/bosh-backup-and-restore-meta/genesis-bosh/bosh.pem
export POSTGRES_PASSWORD

cd src/github.com/pivotal-cf/backup-and-restore-sdk-release
glide install
ginkgo --focus=postgres system_tests