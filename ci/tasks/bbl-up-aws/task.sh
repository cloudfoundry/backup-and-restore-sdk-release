#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

# Provisions a real (non-lite) BOSH director on EC2, for the IAM instance
# profile test: it needs an actual EC2 VM with an instance profile attached.

ci_dir="${PWD}/backup-and-restore-sdk-release-ci/ci"

# bbl names its IAM roles, policies and instance profiles after the env ID,
# and the pipeline role only allows IAM resources named bbr-sdk-*. --name sets
# the full env ID, so make it unique per run: e.g. bbr-sdk-2026-10-22t10-00z-ab12.
env_name="bbr-sdk-$(date -u +%Y-%m-%dt%H-%Mz)-$(openssl rand -hex 2)"

pushd bbl-state > /dev/null
  bbl plan --name "${env_name}"
  bbl --debug up

  eval "$(bbl print-env)"

  # Add the vm_extension the s3-backuper-with-iam-instance-profile manifest uses.
  bosh cloud-config > cloud-config.yml
  bosh -n update-cloud-config cloud-config.yml \
    -o "${ci_dir}/manifests/ops-files/iam-instance-profile.yml" \
    -v iam_instance_profile="${IAM_INSTANCE_PROFILE}"
popd > /dev/null
