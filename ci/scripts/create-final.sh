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

set -ex

VERSION=$(cat version/number)

pushd backup-and-restore-sdk-release
  echo "---
blobstore:
  s3:
    access_key_id: $AWS_ACCESS_KEY_ID
    secret_access_key: $AWS_SECRET_ACCESS_KEY
" > config/private.yml

  bosh -n finalize release ../release-tarball/backup-and-restore-sdk-*.tgz --version "$VERSION"

  mv releases/backup-and-restore-sdk/backup-and-restore-sdk-"${VERSION}".tgz \
    ../backup-and-restore-sdk-final-release-tarball

  git add ./releases
  git add ./.final_builds

  git config --global user.name "Backup & Restore Concourse"
  git config --global user.email "cf-lazarus@pivotal.io"

  git commit -m "Add final release ${VERSION} [ci skip]"
popd

cp -R backup-and-restore-sdk-release/. backup-and-restore-sdk-final-release
