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

export TEST_AWS_ACCESS_KEY_ID="$(bosh-cli int --path=/aws-access-key-id secrets.yml)"
export TEST_AWS_SECRET_ACCESS_KEY="$(bosh-cli int --path=/aws-secret-access-key secrets.yml)"

export TEST_ECS_ACCESS_KEY_ID="$(bosh-cli int --path=/ecs-access-key-id secrets.yml)"
export TEST_ECS_SECRET_ACCESS_KEY="$(bosh-cli int --path=/ecs-secret-access-key secrets.yml)"

ginkgo -trace
