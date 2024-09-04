#!/usr/bin/env bash

[ -z "$DEBUG" ] || set -x

set -euo pipefail
shopt -s nocasematch

################################################################################################
# This script creates a PR to a github repository, from BRANCH to BASE
################################################################################################

# Preventive checks with informative descriptions.
function checks() {
    [[ -z "${GH_TOKEN:-}${GH_ENTERPRISE_TOKEN:-}" ]] && echo -e '\nERROR: required variables ${GH_TOKEN} and ${GH_ENTERPRISE_TOKEN} are missing. At least 1 is required.'
    [[ -z "${BRANCH:-}" ]]   && echo -e '\nERROR: required variable ${BRANCH} is missing. That is the branch you want your PR to be created from.'
    [[ -z "${BASE:-}" ]]     && echo -e '\nERROR: required variable ${BASE} is missing. That is the branch you want your PR to be created to.'
}

export CHECKS="$(checks 2>&1)"
export ERRORS="$(echo "${CHECKS}" | grep "^ERROR: ")"

echo "${CHECKS}"
if [[ -n "${ERRORS}" ]]; then
    exit 1
fi

TITLE="${TITLE:-"[CI Bot] Latest Release Notes"}"

if [ -z "${MESSAGE}" ]; then
  MESSAGE=$(cat <<-EOM
This is an automatically generated Pull Request from the Cryogenics CI Bot.

I have updated the release notes with the latest release and contents.

If this does not look right, please reach out to the #mapbu-cryogenics team.
EOM
)
fi

pushd backup-and-restore-sdk-release
  git checkout "${BRANCH}"

  cmd='gh pr create '
  cmd+="--base '${BASE}' "
  cmd+="--title '${TITLE}' "
  cmd+="--body '${MESSAGE}' "

  if [ -n "$LABELS" ]
  then
    cmd+="--label '${LABELS}' "
  fi

  if "$DRAFT"
  then
    cmd+="--draft "
  fi

  set +e
  output="$(eval ${cmd} 2>&1)"
  pr_exit_status=$?
  set -e

  if [[ "${output}" =~ "a pull request for branch" ]] && [[ "${output}" =~ "already exists" ]]; then
    echo "A PR already exists"
    exit 0
  fi

  if [[ "${output}" =~ "No commits between" ]]; then
    echo "No commits were made between the branches"
    exit 0
  fi
popd

if [[ ! $pr_exit_status -eq 0 ]]; then
  echo "Creating PR failed with: $output";
fi

exit $pr_exit_status

exec "$@"