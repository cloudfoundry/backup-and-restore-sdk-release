#!/bin/bash -eu

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

export BOSH_URL="https://lite-bosh.backup-and-restore.cf-app.com:25555"
export BOSH_CERT_PATH="~/workspace/bosh-backup-and-restore-meta/certs/lite-bosh.backup-and-restore.cf-app.com.crt"
export BOSH_CLIENT="admin"
export BOSH_CLIENT_SECRET=$(lpass show "LiteBoshDirector" --password)
export BOSH_GATEWAY_USER=vcap
export BOSH_GATEWAY_HOST=lite-bosh.backup-and-restore.cf-app.com
export BOSH_GATEWAY_KEY=~/workspace/bosh-backup-and-restore-meta/genesis-bosh/bosh.pem
export POSTGRES_PASSWORD=foo
export MYSQL_PASSWORD=foo

TEST_SUITE=""
if [[ $# -ge 1 ]]; then
    TEST_SUITE=$1
fi

pushd $(dirname $0)/..
    if [[ "${TEST_SUITE}" == "" ]]; then
        ginkgo system_tests -trace
    else
        ginkgo --focus=${TEST_SUITE} system_tests -trace
    fi
popd
