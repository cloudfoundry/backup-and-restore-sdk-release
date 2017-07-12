#!/usr/bin/env bash

set -eu

./bosh-backup-and-restore-meta/unlock-ci.sh
export BOSH_CLIENT
export BOSH_CLIENT_SECRET

bosh-cli --non-interactive -e genesis-bosh.backup-and-restore.cf-app.com \
  --ca-cert "./bosh-backup-and-restore-meta/certs/genesis-bosh.backup-and-restore.cf-app.com.crt" \
  -d ${BOSH_DEPLOYMENT} \
  deploy backup-and-restore-sdk-release/ci/manifests/mysql.yml \
  -v backup-and-restore-sdk-release-version=$(cat release-tarball/version) \
  -v backup-and-restore-sdk-release-url=$(cat release-tarball/url) \
  -v mysql-password=${MYSQL_PASSWORD} \
  -v deployment-name=${BOSH_DEPLOYMENT}
