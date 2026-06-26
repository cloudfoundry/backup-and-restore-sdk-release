#!/usr/bin/env bash
set -euo pipefail
[ -n "${DEBUG:-}" ] && set -x

# Minimal bbl-up for AWS — provisions a real (non-lite) BOSH director on EC2.
# Used for the IAM instance profile system test, which requires an actual
# EC2 instance with an IAM instance profile attached.
# After bbl up, the cloud config is updated to add a vm_extensions entry
# that attaches the IAM instance profile to BOSH-deployed VMs.

pushd "${PWD}/bbl-state"
  bbl plan --iaas aws
  bbl --debug up

  # Add the IAM vm_extension to the cloud config so that the
  # s3-backuper-with-iam-instance-profile manifest can reference it.
  eval "$(bbl print-env)"

  cat > /tmp/iam-vm-extension-ops.yml << OPSEOF
- type: replace
  path: /vm_extensions/-
  value:
    name: s3-backup-iam-instance-profile
    cloud_properties:
      iam_instance_profile: ${BBL_AWS_IAM_INSTANCE_PROFILE}
OPSEOF

  bosh cloud-config > /tmp/current-cloud-config.yml
  bosh int /tmp/current-cloud-config.yml -o /tmp/iam-vm-extension-ops.yml > /tmp/updated-cloud-config.yml
  bosh update-cloud-config /tmp/updated-cloud-config.yml --non-interactive
popd
