#!/usr/bin/env bash
set -euo pipefail
[ -n "${DEBUG:-}" ] && set -x

pushd "${PWD}/bbl-state"
  if [[ ! -f bbl-state.json ]]; then
    echo "No bbl state found; bbl up never completed, nothing to tear down."
    exit 0
  fi

  # Delete all BOSH deployments before tearing down to ensure clean
  # shutdown of any EC2 instances and their attached EBS volumes.
  echo "Pre-cleanup: deleting BOSH deployments before tearing down..."
  bbl_env="$(bbl print-env 2>/dev/null || true)"
  if [[ -n "${bbl_env}" ]]; then
    eval "${bbl_env}" || true
    deployments="$(timeout 60 bosh deployments --json 2>/dev/null | jq -r '.Tables[0].Rows[].name' 2>/dev/null || true)"
    if [[ -n "${deployments}" ]]; then
      echo "${deployments}" | while IFS= read -r dep; do
        echo "Pre-cleanup: deleting deployment '${dep}'..."
        timeout 600 bosh -d "${dep}" delete-deployment --force -n 2>&1 \
          || echo "WARNING: Failed to delete deployment '${dep}'"
      done
    else
      echo "No BOSH deployments found; skipping pre-cleanup."
    fi
  else
    echo "WARNING: Could not source BOSH env; skipping deployment pre-cleanup"
  fi

  bbl --debug down --no-confirm
popd
