#!/usr/bin/env bash

set -e

: "${BOSH_CREDS_SCRIPT:?Variable not set or empty}"
eval "$BOSH_CREDS_SCRIPT"

pushd "/backup-and-restore-sdk-release"
  bosh releases
popd
