#!/bin/bash
set -euo pipefail
set -x

eval "$(bbl print-env --metadata-file cf-deployment-env/metadata)"

bosh cloud-config > cc.yml
bosh update-cloud-config -n cc.yml -o backup-and-restore-sdk-release/ci/tasks/update-cloud-config/update-compilation-vm.yml
