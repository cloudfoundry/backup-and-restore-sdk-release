#!/usr/bin/env bash

set -eux

pushd $(dirname $0)
  ./make_golang_project_spec_file.sh blobstore-backup-restorer blobstore-backup-restore
  ./make_golang_project_spec_file.sh azure-blobstore-backup-restorer azure-blobstore-backup-restore
  ./make_golang_project_spec_file.sh database-backup-restorer database-backup-restore
popd
