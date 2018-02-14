#!/usr/bin/env bash

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

set -eu

./bosh-backup-and-restore-meta/unlock-ci.sh
export BOSH_CLIENT
export BOSH_CLIENT_SECRET
export BOSH_ENVIRONMENT
export BOSH_CA_CERT="./bosh-backup-and-restore-meta/certs/${BOSH_ENVIRONMENT}.crt"

chmod 400 bosh-backup-and-restore-meta/genesis-bosh/bosh.pem

bosh-cli \
  --deployment ${BOSH_DEPLOYMENT} \
  ssh \
  --gw-host=${BOSH_ENVIRONMENT} \
  --gw-user=vcap \
  --gw-private-key=bosh-backup-and-restore-meta/genesis-bosh/bosh.pem \
  -c 'echo -e "hostssl all mutual_tls_user 0.0.0.0/0 cert map=cnmap\nhostssl all ssl_user 0.0.0.0/0 md5\nhost all test_user 0.0.0.0/0 md5" | sudo tee /var/vcap/jobs/postgres/config/pg_hba.conf && sudo /var/vcap/bosh/bin/monit restart postgres && while ! nc -z localhost 5432 </dev/null; do sleep 1; done'