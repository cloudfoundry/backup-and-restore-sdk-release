#!/bin/bash

[ -z "$DEBUG" ] || set -x

set -euo pipefail

: "${GCP_SERVICE_ACCOUNT_KEY:?}"

[ -d env ]

gcloud -q auth activate-service-account --key-file=<(echo "$GCP_SERVICE_ACCOUNT_KEY")
gcloud -q config set project "$(echo "$GCP_SERVICE_ACCOUNT_KEY" | jq -r '.project_id')"
gcloud -q config set compute/zone europe-west1-c

env_name="$(cat env/name)"

rule_name="${env_name}-jumpbox-to-databases"
gcloud compute firewall-rules delete "${rule_name}" --quiet || true
gcloud compute firewall-rules create "${rule_name}" \
        --network="${env_name}-network"     \
        --direction=INGRESS \
        --action=allow \
        --rules=tcp:5432,tcp:3306 \
        --source-tags="${env_name}-jumpbox" \
        --target-tags="${env_name}-internal" \
        --priority=999

cat <<EOF
======================================================================
The following firewall EGRESS rules have been created
----------------------------------------------------------------------

EOF

gcloud compute firewall-rules list --filter="${env_name}-internetless-"