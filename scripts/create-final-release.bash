#!/usr/bin/env bash

set -euxo pipefail

pushd /backup-and-restore-sdk-release
cat << EOF > config/private.yml
---
final_name: backup-and-restore-sdk
blobstore:
  options:
    access_token: ${ACCESS_TOKEN}
EOF

bosh create-release --final --tarball backup-and-restore-sdk.tgz

popd
