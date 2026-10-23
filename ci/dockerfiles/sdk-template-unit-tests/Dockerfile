ARG RUBY_IMAGE=ruby:4.0.9@sha256:1b1e1a97aa1f3bffa480c3e2f9df109f6351d2e8e306a79dc64dfbcf06bb1730
FROM ${RUBY_IMAGE}

ENTRYPOINT /backup-and-restore-sdk-release/scripts/run-template-tests.bash

COPY ./ /backup-and-restore-sdk-release
