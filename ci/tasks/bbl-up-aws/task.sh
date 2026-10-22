#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

ci_dir="${PWD}/backup-and-restore-sdk-release-ci/ci"

# bbl names the environment's AWS resources after its env ID, which --name
# sets, so make it unique per run: e.g. bbr-sdk-2026-10-22t10-00z-ab12.
env_name="bbr-sdk-$(date -u +%Y-%m-%dt%H-%Mz)-$(openssl rand -hex 2)"

pushd bbl-state > /dev/null
  bbl plan --name "${env_name}"

  # The pipeline role can't create IAM resources, so give the director an
  # existing instance profile, and turn off the VPC flow logs, whose role bbl
  # would otherwise create. Nobody reads the flow logs of a VPC that only
  # lives for one run. The AWS provider validates arguments even when count
  # is 0, so the flow logs policy's role needs a valid (unused) name.
  echo "bosh_iam_instance_profile = \"${DIRECTOR_INSTANCE_PROFILE}\"" > vars/iam.tfvars
  cat > terraform/no-flow-logs_override.tf <<'EOF'
resource "aws_flow_log" "bbl" {
  count           = 0
  log_destination = ""
  iam_role_arn    = ""
}

resource "aws_cloudwatch_log_group" "bbl" {
  count = 0
}

resource "aws_iam_role" "flow_logs" {
  count = 0
}

resource "aws_iam_role_policy" "flow_logs" {
  count = 0
  role  = "unused"
}
EOF

  # With BBL_AWS_ASSUME_ROLE, bbl makes both create-env and the director's
  # CPI assume the pipeline role. create-env runs here with the IAM user's
  # keys and needs it. The director runs as DIRECTOR_INSTANCE_PROFILE, which
  # already has the EC2 permissions it needs (and iam:PassRole for the
  # backuper instance profile) but isn't trusted by the pipeline role.
  # So drop the role from the director only, via an override that bbl runs
  # instead of create-director.sh.
  cat > director-without-role-arn.yml <<'EOF'
- type: remove
  path: /instance_groups/name=bosh/properties/aws/role_arn
EOF
  {
    sed '$ s/$/ \\/' create-director.sh
    # shellcheck disable=SC2016 # expanded by bbl when it runs the script
    echo '  -o ${BBL_STATE_DIR}/director-without-role-arn.yml'
  } > create-director-override.sh
  chmod +x create-director-override.sh

  bbl --debug up

  eval "$(bbl print-env)"

  # Add the vm_extension the s3-backuper-with-iam-instance-profile manifest
  # uses. bbl's default vm_type is m4.large, which eu-north-1 doesn't offer
  # ("The requested configuration is currently not supported"), so use
  # m5.large instead.
  cat > default-vm-type.yml <<'EOF'
- type: replace
  path: /vm_types/name=default/cloud_properties/instance_type
  value: m5.large
EOF
  bosh cloud-config > cloud-config.yml
  bosh -n update-cloud-config cloud-config.yml \
    -o "${ci_dir}/manifests/ops-files/iam-instance-profile.yml" \
    -o default-vm-type.yml \
    -v iam_instance_profile="${IAM_INSTANCE_PROFILE}"
popd > /dev/null
