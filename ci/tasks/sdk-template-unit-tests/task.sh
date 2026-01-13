#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

pushd ruby-install
  VERSION=$(cat version)
  tar -xzvf "ruby-install-${VERSION}.tar.gz"
  pushd "ruby-install-${VERSION}"
    make install
  popd
popd

pushd backup-and-restore-sdk-release
  echo 'gem: --no-document' > /etc/gemrc
  NUM_CPUS=$(grep -c ^processor /proc/cpuinfo)
  RUBY_VERSION="$(cat .ruby-version)"
  BUNDLER_VERSION="$(grep 'BUNDLED WITH' -A1 Gemfile.lock | tail -n1 | tr -d '[:blank:]')"
  ruby-install --jobs="${NUM_CPUS}" --cleanup --system ruby "${RUBY_VERSION}" \
    -- --disable-install-rdoc --disable-install-doc

  gem install bundler -v "${BUNDLER_VERSION}"
  bundle install
  bundle exec rspec
popd
