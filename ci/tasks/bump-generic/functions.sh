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

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
REPO_ROOT="$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel)"

function setup_private_blobstore_config() {
    AWS_ACCESS_KEY_ID="$1"
    AWS_SECRET_ACCESS_KEY="$2"

    pushd "${REPO_ROOT}" > /dev/null
    echo "---
    blobstore:
    provider: s3
    options:
        access_key_id: ${AWS_ACCESS_KEY_ID:-}
        secret_access_key: ${AWS_SECRET_ACCESS_KEY:-}
    " > config/private.yml
}


function current_blobs_names() {
    BLOBS_PREFIX="$1"

    pushd "${REPO_ROOT}" >/dev/null
    CURRENT_BLOBS_NAMES="$(bosh blobs | grep "^${BLOBS_PREFIX}/" | cut -f1)"
    for BLOB_NAME in  $CURRENT_BLOBS_NAMES;
    do
        echo "${BLOB_NAME}" | xargs
    done
    popd >/dev/null
}


function download_version() {
    VERSION="$1"
    eval DOWNLOAD_URL="$2"
    eval DOWNLOAD_DESTINATION="$3"

    wget -q -O "${DOWNLOAD_DESTINATION}" "${DOWNLOAD_URL}"
    # TODO: Implement checksum
    realpath "${DOWNLOAD_DESTINATION}"
}


function pick_cadidate_version() {
    CUR_VERSION="$1"
    ALL_VERSIONS="$2"

    if echo "${CUR_VERSION}" | grep -Eo '^[0-9]+$' >/dev/null;
    then # Autobump majors
        AUTOBUMP_PREFIX=""

    elif echo "${CUR_VERSION}" | grep -Eo '^[0-9]+\.[0-9]+$' >/dev/null;
    then # Autobump minors
        AUTOBUMP_PREFIX="$(echo "${CUR_VERSION}" | cut -d '.' -f1)"

    elif echo "${CUR_VERSION}" | grep -Eo '^[0-9]+\.[0-9]+\.[0-9]+$' >/dev/null;
    then # Autobump patches
        AUTOBUMP_PREFIX="$(echo "${CUR_VERSION}" | cut -d '.' -f1,2)"

    else
        echo "Unsupported naming convention: ${CUR_VERSION}"
        exit 1
    fi

    ALL_VERSIONS_SORTED="$(echo "${CUR_VERSION}"$'\n'"${ALL_VERSIONS}" | sort -t "." -k1,1n -k2,2n -k3,3n | uniq)"
    AUTOBUMP_CANDIDATES="$(echo "${ALL_VERSIONS_SORTED}" | grep "^${AUTOBUMP_PREFIX}" || true)"

    # The following sed expression returns all lines after "CUR_VERSION" is found
    NEWER_CANDIDATES="$(echo "${AUTOBUMP_CANDIDATES}" | sed -n '/^'"${CUR_VERSION}"'$/,$p')"
    NEWEST_CANDIDATE="$(echo "${NEWER_CANDIDATES}" | tail -n 1)"

    echo "${NEWEST_CANDIDATE}"
}


function blobs_are_equal() {
    CUR_BLOB_ID="${1}"
    NEW_BLOB_ID="${2}"
    CUR_VERSION="${3}"
    NEW_VERSION="${4}"
    NEW_TARFILE="${5}"

    # TODO: Extend this check to also compare shasum and id
    if [[ "${CUR_VERSION}" == "${NEW_VERSION}" ]];
    then
        return 0
    else
        return 1
    fi
}


function replace_blob() {
    CUR_BLOB_ID="${1}"
    NEW_BLOB_ID="${2}"
    CUR_VERSION="${3}"
    NEW_VERSION="${4}"
    NEW_TARFILE="${5}"

    pushd "${REPO_ROOT}" >/dev/null

    echo "Replacing Postgresql ${CUR_VERSION} with ${NEW_VERSION}"

    # Replace blobstore blob
    bosh remove-blob "--dir=${REPO_ROOT}" "${CUR_BLOB_ID}"
    bosh add-blob "--dir=${REPO_ROOT}" "$(realpath "${NEW_TARFILE}")" "${NEW_BLOB_ID}"
    bosh upload-blobs "--dir=${REPO_ROOT}"

    # Replace references in files
    # Following steps **SHOULD NEVER BE PERFORMED BEFORE** the blobstore replacement commands listed above
    FILES_WITH_REFS="$(grep -rnwl '.' -e "${CUR_BLOB_ID}")"

    for file in $FILES_WITH_REFS;
    do
    sed -i.bak "s#${CUR_BLOB_ID}#${NEW_BLOB_ID}#g" "$file"
    sed -i.bak "s#${CUR_VERSION}#${NEW_VERSION}#g" "$file"
    rm "$file".bak
    done

    popd >/dev/null
}


function committed_changes() {
    BRANCH_NAME="${1:?}"
    COMMIT_MESSAGE="${2:?}"
    COMMIT_USERNAME="${3:?}"
    COMMIT_USEREMAIL="${4:?}"
    GH_USER="${5:?}"
    GH_TOKEN="${6:?}"

    git checkout -b "${BRANCH_NAME}"
    echo "Pushing updates to the configured branch '${BRANCH_NAME}'"

    git config user.name "${COMMIT_USERNAME}"
    git config user.email "${COMMIT_USEREMAIL}"

    if git commit -m "${COMMIT_MESSAGE}"; then
        echo "${COMMIT_MESSAGE}"

        origin_url="$(git remote get-url origin | grep -Eo 'github.com.*' | sed 's/github.com:/github.com\//g')"
        git remote set-url --push autobump "https://${GH_USER}:${GH_TOKEN}@${origin_url}/"
        git push -u autobump "${BRANCH_NAME}" --force
        git remote remove autobump
        return 1
    else
        echo "No change to be committed"
        return 0
    fi
}


function create_pr() {
    PR_BRANCH="${1:?}"
    PR_BASE="${2:?}"
    PR_TITLE="${3:?}"
    PR_MESSAGE="${4:?}"
    PR_LABELS="${5:?}"
    GH_USER="${6:?}"
    GH_TOKEN="${7:?}"

    pushd "${REPO_ROOT}"
    git checkout "${PR_BRANCH}"

    set +e
    output="$(gh pr create \
        --base "${PR_BASE}" \
        --title "${PR_TITLE}" \
        --body "${PR_MESSAGE}" \
        --label "${PR_LABELS}" 2>&1)"
    pr_exit_status=$?
    set -e

    if [[ "${output}" =~ "No commits between" ]]; then
        echo "No commits were made between the branches"
        exit 0
    fi
    popd

    exit $pr_exit_status
}


function safely_expand_variables() {
    TEXT_TO_EXPAND="$1"

    export BLOBS_PREFIX
    export PREV_VERSION
    export NEW_VERSION
    export VERSIONS_URL

    echo "${TEXT_TO_EXPAND}" | envsubst '${BLOBS_PREFIX} ${PREV_VERSION} ${NEW_VERSION} ${VERSIONS_URL}'
}
