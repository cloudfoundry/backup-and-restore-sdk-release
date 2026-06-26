#!/usr/bin/env bash
set -euo pipefail
[ -n "${DEBUG:-}" ] && set -x

# Fallback: when bbl down fails because CleanUpDirector cannot reach the BOSH
# director API (e.g. the director process is unhealthy due to a disk mount
# failure inside the VM), run the individual deletion steps directly.
#
# bbl down sequence:
#   1. CleanUpDirector  – connects to director API, runs "bosh clean-up --all"
#   2. DeleteDirector   – runs delete-director-override.sh (bosh delete-env via GCP CPI)
#   3. DeleteJumpbox    – runs delete-jumpbox-override.sh  (bosh delete-env via GCP CPI)
#   4. terraform destroy
#
# Steps 2-4 use the GCP API directly and do not require a healthy director.
# Bypassing step 1 allows them to run even when the director is unreachable.

bbl_direct_destroy() {
  local bbl_state_dir="${PWD}"

  export BBL_STATE_DIR="${bbl_state_dir}"
  export BBL_GCP_PROJECT_ID
  BBL_GCP_PROJECT_ID=$(echo "${BBL_GCP_SERVICE_ACCOUNT_KEY}" | jq -r '.project_id')
  export BBL_GCP_ZONE
  BBL_GCP_ZONE=$(grep -m1 '^zone:' vars/director-vars-file.yml 2>/dev/null \
    | awk '{print $2}' | tr -d '"' || true)
  BBL_GCP_ZONE="${BBL_GCP_ZONE:-${BBL_GCP_REGION}-a}"

  local sa_key_path
  sa_key_path=$(mktemp /tmp/gcp-sa-key.XXXXX.json)
  printf '%s' "${BBL_GCP_SERVICE_ACCOUNT_KEY}" > "${sa_key_path}"
  export BBL_GCP_SERVICE_ACCOUNT_KEY_PATH="${sa_key_path}"

  sh delete-director-override.sh \
    || echo "WARNING: delete-director-override.sh failed; director VM may already be absent"

  sh delete-jumpbox-override.sh \
    || echo "WARNING: delete-jumpbox-override.sh failed; jumpbox VM may already be absent"

  local var_args=()
  for f in "${bbl_state_dir}"/vars/*.tfvars; do
    var_args+=("-var-file=${f}")
  done
  pushd "${bbl_state_dir}/terraform"
    terraform init
    terraform destroy -auto-approve \
      "-state=${bbl_state_dir}/vars/terraform.tfstate" \
      "-var=credentials=${sa_key_path}" \
      "${var_args[@]}"
  popd

  rm -f "${sa_key_path}"
}

pushd "${PWD}/bbl-state"
  if [[ ! -f bbl-state.json ]]; then
    echo "No bbl state found; bbl up never completed, nothing to tear down."
    exit 0
  fi

  # Delete all BOSH deployments before tearing down the environment.
  # In BOSH-lite, warden container persistent volumes are bind-mounted from
  # the director's persistent disk. If bind mounts are still active when
  # bbl down tries to unmount the disk, umount fails with "target is busy"
  # (exit 32) and bosh delete-env aborts after a 10-minute timeout.
  echo "Pre-cleanup: sourcing BOSH env to delete deployments before tearing down..."
  bbl_env="$(bbl print-env 2>/dev/null || true)"
  if [[ -n "${bbl_env}" ]]; then
    eval "${bbl_env}" || true
    deployments="$(timeout 60 bosh deployments --json 2>/dev/null | jq -r '.Tables[0].Rows[].name' 2>/dev/null || true)"
    if [[ -n "${deployments}" ]]; then
      echo "${deployments}" | while IFS= read -r dep; do
        echo "Pre-cleanup: deleting deployment '${dep}' to release warden bind mounts..."
        timeout 600 bosh -d "${dep}" delete-deployment --force -n 2>&1 \
          || echo "WARNING: Failed to delete deployment '${dep}'; bind mounts may still be active"
      done
    else
      echo "No BOSH deployments found; skipping pre-cleanup."
    fi
  else
    echo "WARNING: Could not source BOSH env; skipping deployment pre-cleanup"
  fi

  if bbl --debug down --no-confirm; then
    exit 0
  fi

  echo "bbl down failed; falling back to direct GCP cleanup (bypassing CleanUpDirector)"
  bbl_direct_destroy
popd
