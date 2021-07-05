#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

pushd "${SCRIPT_DIR}" > /dev/null
CURRENT_VERSION="$(./current_version.sh)"
STABLE_RELEASES="$(./list_all_stable_releases.sh)"
SORTED_RELEASES="$(echo "${CURRENT_VERSION}"$'\n'"${STABLE_RELEASES}" | sort -t "." -k1,1n -k2,2n -k3,3n | uniq)"

# The following sed expression returns all lines after match
NEW_RELEASES="$(echo "${SORTED_RELEASES}" | sed -n '/^'"${CURRENT_VERSION}"'$/,$p')"
echo "${NEW_RELEASES}"
popd > /dev/null