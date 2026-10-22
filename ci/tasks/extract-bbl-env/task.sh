#!/usr/bin/env bash
set -euo pipefail

pushd bbl-state > /dev/null
  director_address="$(bbl director-address)"
  director_username="$(bbl director-username)"
  director_password="$(bbl director-password)"
  bbl director-ca-cert > ../bosh-env/ca.crt
  jumpbox_address="$(bbl jumpbox-address)"
  jumpbox_private_key="$(bbl ssh-key)"
popd > /dev/null

# bbl's jumpbox user is always `jumpbox`.
cat > bosh-env/env.sh <<EOF
bosh_env_dir="\$(cd "\$(dirname "\${BASH_SOURCE[0]}")" && pwd)"
export BOSH_ENVIRONMENT=$(printf '%q' "${director_address}")
export BOSH_CLIENT=$(printf '%q' "${director_username}")
export BOSH_CLIENT_SECRET=$(printf '%q' "${director_password}")
export BOSH_CA_CERT="\${bosh_env_dir}/ca.crt"
export BOSH_GW_HOST=$(printf '%q' "${jumpbox_address%:*}")
export BOSH_GW_USER=jumpbox
export BOSH_GW_PRIVATE_KEY=$(printf '%q' "${jumpbox_private_key}")
EOF
