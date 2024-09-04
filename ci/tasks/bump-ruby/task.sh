#!/usr/bin/env bash

[ -z "$DEBUG" ] || set -x

set -euo pipefail

[ -d backup-and-restore-sdk-release ]

cd backup-and-restore-sdk-release

EXISTING_RUBY_VERSION="$(cat .ruby-version)"
NEW_RUBY_VERSION="$(ruby --version | cut -d ' ' -f2)"
echo "$NEW_RUBY_VERSION" > .ruby-version

bundle update --ruby
bundle update --bundler

git config user.name "${GIT_USERNAME}"
git config user.email "${GIT_EMAIL}"

git add .ruby-version
git add Gemfile.lock
git commit -m "Bump Ruby from ${EXISTING_RUBY_VERSION} to ${NEW_RUBY_VERSION}"