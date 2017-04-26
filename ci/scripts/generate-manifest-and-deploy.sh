#!/usr/bin/env bash

set -eu

./bosh-backup-and-restore-meta/unlock-ci.sh
export BOSH_CLIENT
export BOSH_CLIENT_SECRET

bosh-cli --non-interactive -e genesis-bosh.backup-and-restore.cf-app.com \
  --ca-cert "./bosh-backup-and-restore-meta/certs/genesis-bosh.backup-and-restore.cf-app.com.crt" \
  -d ${BOSH_DEPLOYMENT} \
  deploy database-backup-and-restore-release/ci/manifests/postgres-dev.yml \
  -v database-backup-and-restore-release-version=$(cat release-tarball/version) \
  -v database-backup-and-restore-release-url=$(cat release-tarball/url) \
  -v postgres-password=${POSTGRES_PASSWORD} \
  -v deployment-name=${BOSH_DEPLOYMENT}