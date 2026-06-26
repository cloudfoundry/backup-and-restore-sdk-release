#!/usr/bin/env bash
set -euo pipefail
[ -n "${DEBUG:-}" ] && set -x

bbl_up() {
  bbl plan
  rm -rf bosh-deployment
  cp -rfp "${bosh_deployment}" .
  cp -R "${bosh_bootloader}/plan-patches/bosh-lite-gcp/." .

  # GCP labels applied to the BOSH director VM (created by bosh create-env).
  # Using individual key paths (labels?/key?) so that any labels already present
  # in cloud_properties are preserved rather than replaced.
  cat > gcp-labels-director.yml << OPSEOF
---
- type: replace
  path: /resource_pools/name=vms/cloud_properties/labels?/pipeline?
  value: sdk
- type: replace
  path: /resource_pools/name=vms/cloud_properties/labels?/pipeline-job?
  value: ${PIPELINE_JOB}
OPSEOF

  # GCP labels applied to the jumpbox VM (created by bosh create-env).
  cat > gcp-labels-jumpbox.yml << OPSEOF
---
- type: replace
  path: /resource_pools/name=vms/cloud_properties/labels?/pipeline?
  value: sdk
- type: replace
  path: /resource_pools/name=vms/cloud_properties/labels?/pipeline-job?
  value: ${PIPELINE_JOB}
OPSEOF

  # Inject ops-file flags into create-director-override.sh.
  sed '$ s/$/ \\/' create-director-override.sh > /tmp/create-director-override.sh
  printf ' -o ${BBL_STATE_DIR}/bosh-deployment/bbr.yml \\\n' >> /tmp/create-director-override.sh
  if [[ "${WARDEN_CONTAINERS_USE_SYSTEMD:-true}" == "false" ]]; then
    cat > bosh-lite-jammy-containers.yml << OPSEOF
---
- type: replace
  path: /instance_groups/name=bosh/properties/warden_cpi/start_containers_with_systemd?
  value: false
OPSEOF
    printf ' -o ${BBL_STATE_DIR}/bosh-lite-jammy-containers.yml \\\n' >> /tmp/create-director-override.sh
  fi
  if [[ "${DISABLE_HM_RESURRECTOR:-false}" == "true" ]]; then
    cat > bosh-disable-resurrector.yml << OPSEOF
---
- type: replace
  path: /instance_groups/name=bosh/properties/hm/resurrector_enabled
  value: false
OPSEOF
    printf ' -o ${BBL_STATE_DIR}/bosh-disable-resurrector.yml \\\n' >> /tmp/create-director-override.sh
  fi

  # bosh-lite.yml pins os-conf v18 which only has basic jobs (sysctl, disable_agent, etc.).
  # Upgrade to v23 so we can use pre-start-script and the iptables job, both of which
  # were added in later releases and are required for the director VM configuration below.
  cat > bosh-os-conf-upgrade.yml << 'OPSEOF'
---
- path: /releases/name=os-conf
  type: replace
  value:
    name: os-conf
    sha1: "sha256:efcf30754ce4c5f308aedab3329d8d679f5967b2a4c3c453204c7cb10c7c5ed9"
    url: https://bosh.io/d/github.com/cloudfoundry/os-conf-release?v=23.0.0
    version: "23.0.0"
OPSEOF
  printf ' -o ${BBL_STATE_DIR}/bosh-os-conf-upgrade.yml \\\n' >> /tmp/create-director-override.sh

  # Noble warden containers share the host kernel's inotify limits. With 15+
  # simultaneous warden containers each running systemd (and apps running Envoy),
  # the default fs.inotify.max_user_instances=128 is quickly exhausted, causing
  # systemd and Envoy to abort with "inotify_fd_ >= 0 assert failure". Setting a
  # high limit ensures every process can create inotify watches without contention.
  cat > bosh-inotify-limits.yml << OPSEOF
---
- path: /instance_groups/name=bosh/jobs/-
  type: replace
  value:
    name: sysctl
    release: os-conf
    properties:
      sysctl:
      - fs.inotify.max_user_instances=65536
      - fs.inotify.max_user_watches=1048576
OPSEOF
  printf ' -o ${BBL_STATE_DIR}/bosh-inotify-limits.yml \\\n' >> /tmp/create-director-override.sh

  # Enable IP forwarding and permissive FORWARD rules so guardian container
  # networking continues to work if anything resets the policy.
  cat > bosh-forward-iptables.yml << 'OPSEOF'
---
- path: /instance_groups/name=bosh/jobs/-
  type: replace
  value:
    name: pre-start-script
    release: os-conf
    properties:
      script: |
        #!/bin/bash
        sysctl -w net.ipv4.ip_forward=1 || echo 1 > /proc/sys/net/ipv4/ip_forward || true
        modprobe nf_conntrack 2>/dev/null || true
        iptables -P FORWARD ACCEPT || true
        iptables -C FORWARD -s 10.0.0.0/24 -d 10.244.0.0/16 -j ACCEPT 2>/dev/null || \
          iptables -I FORWARD 1 -s 10.0.0.0/24 -d 10.244.0.0/16 -j ACCEPT || true
        iptables -C FORWARD -s 10.244.0.0/16 -d 10.0.0.0/24 \
          -m state --state ESTABLISHED,RELATED -j ACCEPT 2>/dev/null || \
          iptables -I FORWARD 2 -s 10.244.0.0/16 -d 10.0.0.0/24 \
          -m state --state ESTABLISHED,RELATED -j ACCEPT || true
OPSEOF
  printf ' -o ${BBL_STATE_DIR}/bosh-forward-iptables.yml \\\n' >> /tmp/create-director-override.sh
  printf ' -o ${BBL_STATE_DIR}/gcp-labels-director.yml\n' >> /tmp/create-director-override.sh
  cp /tmp/create-director-override.sh create-director-override.sh
  chmod +x create-director-override.sh

  # Inject labels ops-file flag into create-jumpbox-override.sh.
  sed '$ s/$/ \\/' create-jumpbox-override.sh > /tmp/create-jumpbox-override.sh
  printf ' -o ${BBL_STATE_DIR}/gcp-labels-jumpbox.yml\n' >> /tmp/create-jumpbox-override.sh
  cp /tmp/create-jumpbox-override.sh create-jumpbox-override.sh
  chmod +x create-jumpbox-override.sh

  # Add GCP labels to the Terraform-managed external IP resources.
  cat > terraform/labels_override.tf << EOF
resource "google_compute_address" "jumpbox-ip" {
  labels = {
    pipeline     = "sdk"
    pipeline-job = "${PIPELINE_JOB}"
  }
}

resource "google_compute_address" "bosh-director-ip" {
  labels = {
    pipeline     = "sdk"
    pipeline-job = "${PIPELINE_JOB}"
  }
}
EOF

  # Increase short_env_id to 32 chars to avoid firewall rule name collisions
  # between concurrent environments sharing the same date prefix.
  sed -i 's/min(20, length(var.env_id))/min(32, length(var.env_id))/' terraform/bosh-lite.tf

  bbl --debug up
}

bosh_deployment="$PWD/bosh-deployment"
bosh_bootloader="$PWD/bosh-bootloader"

pushd "${PWD}/bbl-state"
  bbl_up

  # Configure BOSH DNS to use public recursors.
  # GCP's link-local metadata DNS (169.254.169.254) is not routable from
  # warden containers, causing SERVFAIL for external hostnames.
  eval "$(bbl print-env)"
  cat > /tmp/dns-recursors-ops.yml << 'OPSEOF'
- type: replace
  path: /addons/name=bosh-dns/jobs/name=bosh-dns/properties/recursors?
  value:
  - 8.8.8.8
  - 8.8.4.4
OPSEOF
  bosh runtime-config --name dns > /tmp/current-dns-rc.yml
  bosh int /tmp/current-dns-rc.yml -o /tmp/dns-recursors-ops.yml > /tmp/modified-dns-rc.yml
  bosh update-runtime-config /tmp/modified-dns-rc.yml --name dns --non-interactive

popd
