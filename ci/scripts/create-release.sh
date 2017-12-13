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

set -e
set -x

export VERSION=$(cat version/number)

if [ -z "$RELEASE_NAME" ]; then
  export RELEASE_NAME="backup-and-restore-sdk"
fi

pushd backup-and-restore-sdk-release
  bosh-cli create-release \
    --version $VERSION \
    --name=${RELEASE_NAME} \
    --tarball=../backup-and-restore-sdk-release-build/${RELEASE_NAME}-$VERSION.tgz --force
popd
