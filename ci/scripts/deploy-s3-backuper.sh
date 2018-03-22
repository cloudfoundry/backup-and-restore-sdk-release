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

bosh-cli --non-interactive \
  --deployment ${BOSH_DEPLOYMENT} \
  deploy "backup-and-restore-sdk-release/ci/manifests/${MANIFEST_NAME}" \
  --var=deployment-name=${BOSH_DEPLOYMENT} \
  --var=backup-and-restore-sdk-release-version=$(cat release-tarball/version) \
  --var=backup-and-restore-sdk-release-url=$(cat release-tarball/url) \
  --var=aws-access-key-id=${AWS_ACCESS_KEY_ID} \
  --var=aws-secret-access-key=${AWS_SECRET_ACCESS_KEY} \
  --var=s3-bucket-name=${S3_BUCKET_NAME} \
  --var=s3-cloned-bucket-name=${S3_CLONED_BUCKET_NAME} \
  --var=s3-region=${S3_REGION} \
  --var=s3-cloned-bucket-region=${S3_CLONED_BUCKET_REGION} \
  --var=s3-unversioned-bucket-name-for-versioned-backuper=${S3_UNVERSIONED_BUCKET_NAME_FOR_VERSIONED_BACKUPER} \
  --var=s3-unversioned-bucket-region-for-versioned-backuper=${S3_UNVERSIONED_BUCKET_REGION_FOR_VERSIONED_BACKUPER} \
  --var=s3-unversioned-bucket-name=${S3_UNVERSIONED_BUCKET_NAME} \
  --var=s3-unversioned-bucket-region=${S3_UNVERSIONED_BUCKET_REGION} \
  --var=s3-unversioned-backup-bucket-name=${S3_UNVERSIONED_BACKUP_BUCKET_NAME} \
  --var=s3-unversioned-backup-bucket-region=${S3_UNVERSIONED_BACKUP_BUCKET_REGION} \
  --var=minio-access-key=${MINIO_ACCESS_KEY} \
  --var=minio-secret-key=${MINIO_SECRET_KEY}
