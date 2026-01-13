#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

[ -d backup-and-restore-sdk-release ]

cd backup-and-restore-sdk-release

EXISTING_RUBY_VERSION="$(cat .ruby-version)"
NEW_RUBY_VERSION="$(ruby --version | cut -d ' ' -f2)"
echo "$NEW_RUBY_VERSION" > .ruby-version

bundle update --ruby
bundle update --bundler

if [ -z "$(git status .ruby-version Gemfile.lock --short)" ]; then
  echo "Nothing to update"
  exit 0
fi

git config user.name "${GIT_USERNAME}"
git config user.email "${GIT_EMAIL}"

git add .ruby-version
git add Gemfile.lock
git commit -m "Bump Ruby from ${EXISTING_RUBY_VERSION} to ${NEW_RUBY_VERSION}"