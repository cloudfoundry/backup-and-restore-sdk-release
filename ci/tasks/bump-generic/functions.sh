#!/usr/bin/env bash
set -euo pipefail

# Sane default values for customization variables
GH_USER=${GH_USER:-'Cryogenics-CI'}
COMMIT_USERNAME=${COMMIT_USERNAME:-'Cryogenics CI Bot'}
COMMIT_USEREMAIL=${COMMIT_USEREMAIL:-'mapbu-cryogenics@groups.vmware.com'}
COMMIT_MESSAGE=${COMMIT_MESSAGE:-'Bump ${BLOBS_PREFIX} from ${PREV_VERSION} to ${NEW_VERSION}'}

PR_LABELS=${PR_LABELS:-''} # No labels by default
PR_TITLE=${PR_TITLE:-'Bump ${BLOBS_PREFIX} from ${PREV_VERSION} to ${NEW_VERSION}'}
PR_MESSAGE=${PR_MESSAGE:-'
This is an automatically generated Pull Request from the Cryogenics CI Bot.

I have detected a new version of [${BLOBS_PREFIX}](${VERSIONS_URL}) and automatically bumped
this package to benefit from the latest changes.

If this does not look right, please reach out to the [#mapbu-cryogenics](https://vmware.slack.com/archives/C01DXEYRKRU) team.
'}

REPO_ROOT="$(git -C "$(realpath "$(dirname "${AUTOBUMP_DESCRIPTOR}")")" rev-parse --show-toplevel)"

function setup_private_blobstore_config() {
    local AWS_ACCESS_KEY_ID="$1"
    local AWS_SECRET_ACCESS_KEY="$2"

    pushd "${REPO_ROOT}" > /dev/null
    echo "---
blobstore:
  provider: s3
  options:
    access_key_id: ${AWS_ACCESS_KEY_ID:-}
    secret_access_key: ${AWS_SECRET_ACCESS_KEY:-}
" > config/private.yml
    popd >/dev/null
}


function current_blobs_names() {
    local BLOBS_PREFIX="$1"

    pushd "${REPO_ROOT}" >/dev/null
    local CURRENT_BLOBS_NAMES="$(bosh blobs | grep "^${BLOBS_PREFIX}/" | cut -f1)"
    for BLOB_NAME in  $CURRENT_BLOBS_NAMES;
    do
        echo "${BLOB_NAME}" | xargs
    done
    popd >/dev/null
}


function download_version() {
    local VERSION="$1"
    eval DOWNLOAD_URL="$2"
    eval DOWNLOAD_DESTINATION="$3"

    wget -q -O "${DOWNLOAD_DESTINATION}" "${DOWNLOAD_URL}"
    # TODO: Implement checksum
    realpath "${DOWNLOAD_DESTINATION}"
}


function pick_cadidate_version() {
    local CUR_VERSION="$1"
    local ALL_VERSIONS="$2"

    if echo "${CUR_VERSION}" | grep -Eo '^[0-9]+$' >/dev/null;
    then # Autobump majors
        local AUTOBUMP_PREFIX=""

    elif echo "${CUR_VERSION}" | grep -Eo '^[0-9]+\.[0-9]+$' >/dev/null;
    then # Autobump minors
        local AUTOBUMP_PREFIX="$(echo "${CUR_VERSION}" | cut -d '.' -f1)"

    elif echo "${CUR_VERSION}" | grep -Eo '^[0-9]+\.[0-9]+\.[0-9]+$' >/dev/null;
    then # Autobump patches
        local AUTOBUMP_PREFIX="$(echo "${CUR_VERSION}" | cut -d '.' -f1,2)"

    else
        echo "Unsupported naming convention: ${CUR_VERSION}"
        exit 1
    fi

    local ALL_VERSIONS_SORTED="$(echo "${CUR_VERSION}"$'\n'"${ALL_VERSIONS}" | sort -t "." -k1,1n -k2,2n -k3,3n | uniq)"
    local AUTOBUMP_CANDIDATES="$(echo "${ALL_VERSIONS_SORTED}" | grep "^${AUTOBUMP_PREFIX}" || true)"

    # The following sed expression returns all lines after "CUR_VERSION" is found
    local NEWER_CANDIDATES="$(echo "${AUTOBUMP_CANDIDATES}" | sed -n '/^'"${CUR_VERSION}"'$/,$p')"
    local NEWEST_CANDIDATE="$(echo "${NEWER_CANDIDATES}" | tail -n 1)"

    echo "${NEWEST_CANDIDATE}"
}


function blobs_are_equal() {
    local CUR_BLOB_ID="${1}"
    local NEW_BLOB_ID="${2}"
    local CUR_VERSION="${3}"
    local NEW_VERSION="${4}"

    # TODO: Extend this check to also compare shasum and id
    if [[ "${CUR_VERSION}" == "${NEW_VERSION}" ]];
    then
        return 0
    else
        return 1
    fi
}


function replace_blob() {
    local CUR_BLOB_ID="${1}"
    local NEW_BLOB_ID="${2}"
    local CUR_VERSION="${3}"
    local NEW_VERSION="${4}"
    local NEW_TARFILE="${5}"
    local BLOBS_PREFIX="${6}"

    pushd "${REPO_ROOT}" >/dev/null

    echo "Replacing ${BLOBS_PREFIX} ${CUR_VERSION} with ${NEW_VERSION}"

    # Replace blobstore blob
    bosh remove-blob "--dir=${REPO_ROOT}" "${CUR_BLOB_ID}"
    bosh add-blob "--dir=${REPO_ROOT}" "$(realpath "${NEW_TARFILE}")" "${NEW_BLOB_ID}"
    bosh upload-blobs "--dir=${REPO_ROOT}"

    # Replace references in files
    # Following steps **SHOULD NEVER BE PERFORMED BEFORE** the blobstore replacement commands listed above
    local FILES_WITH_REFS="$(grep -rnwl '.' -e "${CUR_VERSION}" | grep "${BLOBS_PREFIX}")"

    for file in $FILES_WITH_REFS;
    do
    sed -i.bak "s#${CUR_BLOB_ID}#${NEW_BLOB_ID}#g" "$file"
    sed -i.bak "s#${CUR_VERSION}#${NEW_VERSION}#g" "$file"
    rm "$file".bak
    done

    popd >/dev/null
}


function committed_changes() {
    local BRANCH_NAME="${1:?}"
    local COMMIT_MESSAGE="${2:?}"
    local COMMIT_USERNAME="${3:?}"
    local COMMIT_USEREMAIL="${4:?}"
    local GH_USER="${5:?}"
    local GH_TOKEN="${6:?}"

    pushd "${REPO_ROOT}" >/dev/null

    git checkout -b "${BRANCH_NAME}"
    echo "Pushing updates to the configured branch '${BRANCH_NAME}'"

    git config user.name "${COMMIT_USERNAME}"
    git config user.email "${COMMIT_USEREMAIL}"
    git add -A
    if git commit -m "${COMMIT_MESSAGE}"; then
        echo "${COMMIT_MESSAGE}"

        local origin_url="$(git remote get-url origin | grep -Eo 'github.com.*' | sed 's/github.com:/github.com\//g')"
        git remote add autobump "https://${GH_USER}:${GH_TOKEN}@${origin_url}/"
        git push -u autobump "${BRANCH_NAME}" --force
        git remote remove autobump
        return 0
    else
        echo "No change to be committed"
        return 1
    fi

    popd >/dev/null
}


function create_pr() {
    local PR_BRANCH="${1:?}"
    local PR_BASE="${2:?}"
    local PR_TITLE="${3:?}"
    local PR_MESSAGE="${4:?}"
    local PR_LABELS="${5:?}"
    local GH_USER="${6:?}"
    local GH_TOKEN="${7:?}"

    pushd "${REPO_ROOT}"
    git checkout "${PR_BRANCH}"

    local output="$(gh pr create \
        --base "${PR_BASE}" \
        --title "${PR_TITLE}" \
        --body "${PR_MESSAGE}" \
        --label "${PR_LABELS}" \
        --head "${PR_BRANCH}" 2>&1)"

    if [[ "${output}" =~ "No commits between" ]]; then
        echo "No commits were made between the branches"
    fi
    popd
}


function safely_expand_variables() {
    local TEXT_TO_EXPAND="$1"

    export BLOBS_PREFIX
    export PREV_VERSION
    export NEW_VERSION
    export VERSIONS_URL

    echo "${TEXT_TO_EXPAND}" | envsubst '${BLOBS_PREFIX} ${PREV_VERSION} ${NEW_VERSION} ${VERSIONS_URL}'
}


function callback_defined() {
    CALLBACK_NAME="${1}"

    if [[ "$(type -t "${CALLBACK_NAME}")" = "function" ]];
    then
        return 0
    else
        return 1
    fi
}