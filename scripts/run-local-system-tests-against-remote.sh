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

lpass show Shared-PCF-Backup-and-Restore/concourse-secrets --notes > \
  secrets.yml

export BOSH_ENVIRONMENT="https://lite-bosh.backup-and-restore.cf-app.com:25555"
export BOSH_CA_CERT="~/workspace/bosh-backup-and-restore-meta/certs/lite-bosh.backup-and-restore.cf-app.com.crt"
export BOSH_CLIENT="admin"
export BOSH_CLIENT_SECRET=$(lpass show "LiteBoshDirector" --password)
export BOSH_GW_USER=vcap
export BOSH_GW_HOST=lite-bosh.backup-and-restore.cf-app.com
export BOSH_GW_PRIVATE_KEY=~/workspace/bosh-backup-and-restore-meta/genesis-bosh/bosh.pem
export POSTGRES_PASSWORD=postgres_password
export MYSQL_PASSWORD=mysql_password
export BOSH_DEPLOYMENT="s3-backuper"
export AWS_TEST_BUCKET_NAME="bbr-system-test-bucket"
export AWS_TEST_CLONE_BUCKET_NAME="bbr-system-test-bucket-clone"
export AWS_TEST_BUCKET_REGION="eu-west-1"
export AWS_TEST_CLONE_BUCKET_REGION="eu-central-1"
export AWS_ACCESS_KEY_ID="$(bosh-cli int --path=/aws-access-key-id secrets.yml)"
export AWS_SECRET_ACCESS_KEY="$(bosh-cli int --path=/aws-secret-access-key secrets.yml)"

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
