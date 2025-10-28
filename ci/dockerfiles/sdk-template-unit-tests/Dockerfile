ARG RUBY_VERSION
FROM ruby:$RUBY_VERSION

ENTRYPOINT /backup-and-restore-sdk-release/scripts/run-template-tests.bash

COPY ./ /backup-and-restore-sdk-release
