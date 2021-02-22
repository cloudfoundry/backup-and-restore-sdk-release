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

ssh_proxy_key="$(mktemp)"
echo -e "${BOSH_GW_PRIVATE_KEY}" > "$ssh_proxy_key"
chmod 0400 "$ssh_proxy_key"

export SSH_PROXY_HOST="$BOSH_GW_HOST"
export SSH_PROXY_USER="$BOSH_GW_USER"
export SSH_PROXY_KEY_FILE="$ssh_proxy_key"
export BOSH_GW_PRIVATE_KEY="$ssh_proxy_key"

export BOSH_ALL_PROXY="ssh+socks5://${BOSH_GW_USER}@${BOSH_GW_HOST}:22?private-key=${ssh_proxy_key}"

cd backup-and-restore-sdk-release/src/database-backup-restore
ginkgo -mod vendor -v -r "system_tests/${TEST_SUITE_NAME}" -trace
