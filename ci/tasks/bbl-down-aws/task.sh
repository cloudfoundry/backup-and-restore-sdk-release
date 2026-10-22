#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

pushd bbl-state > /dev/null
  if [[ ! -f bbl-state.json ]]; then
    echo "No bbl state found; bbl up never got far enough to create anything."
    exit 0
  fi

  # Delete deployments first, so their EC2 instances and EBS volumes don't
  # outlive the director.
  if bbl_env="$(bbl print-env 2> /dev/null)"; then
    eval "${bbl_env}"
    deployments="$(timeout 60 bosh deployments --json 2> /dev/null | jq -r '.Tables[0].Rows[].name' || true)"
    for deployment in ${deployments}; do
      echo "Deleting deployment ${deployment}"
      timeout 600 bosh -n -d "${deployment}" delete-deployment --force \
        || echo "WARNING: could not delete deployment ${deployment}" >&2
    done
  else
    echo "WARNING: could not read the BOSH env; not deleting deployments first" >&2
  fi

  bbl --debug down --no-confirm
popd > /dev/null
