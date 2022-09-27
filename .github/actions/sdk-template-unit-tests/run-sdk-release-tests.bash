#!/usr/bin/env bash

set -e

pushd "/backup-and-restore-sdk-release"
  bundle
  bundle exec rspec
popd
