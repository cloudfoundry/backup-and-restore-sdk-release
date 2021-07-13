#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

# Check mandatory params without defaults
: "${GH_TOKEN:?}"
: "${PR_BASE:?}"
: "${AWS_ACCESS_KEY_ID:?}"
: "${AWS_SECRET_ACCESS_KEY:?}"

AUTOBUMP_DESCRIPTOR="$1"
source "${AUTOBUMP_DESCRIPTOR}"

# Check params coming from AUTOBUMP_DESCRIPTOR
: "${BLOBS_PREFIX:?}"
: "${VERSIONS_URL:?}"
: "${DOWNLOAD_URL:?}"
: "${DOWNLOADED_FILENAME:?}"
: "${ALL_VERSIONS:?}"

source "${SCRIPT_DIR}/functions.sh"


COMMIT_SAVEPOINT="$(git rev-parse HEAD)"
for BLOB_ID in $(current_blobs_names "${BLOBS_PREFIX}");
do
    setup_private_blobstore_config "${AWS_ACCESS_KEY_ID}" "${AWS_SECRET_ACCESS_KEY}"

    PREV_VERSION="$(echo "${BLOB_ID}" | grep -Eo '[0-9]+(\.[0-9]+)*')"
    NEW_VERSION="$(pick_cadidate_version "${PREV_VERSION}" "${ALL_VERSIONS}")"
    NEW_TARFILE="$(download_version "${NEW_VERSION}" "${DOWNLOAD_URL}" "${SDK_ROOT}/${DOWNLOADED_FILENAME}")"
    NEW_BLOB_ID="$(echo "${BLOB_ID}" | grep "${PREV_VERSION}" | sed "s/${PREV_VERSION}/${NEW_VERSION}/")"

    if blobs_are_equal "${BLOB_ID}" "${NEW_BLOB_ID}" "${PREV_VERSION}" "${NEW_VERSION}" "${NEW_TARFILE}";
    then
        echo "${BLOB_ID} is up-to-date"
    else
        replace_blob "${BLOB_ID}" "${NEW_BLOB_ID}" "${PREV_VERSION}" "${NEW_VERSION}" "${NEW_TARFILE}"
        rm -f "${NEW_TARFILE}"

        COMMIT_MESSAGE="$(safely_expand_variables "${COMMIT_MESSAGE}")"
        PR_MESSAGE="$(safely_expand_variables "${PR_MESSAGE}")"
        PR_TITLE="$(safely_expand_variables "${PR_TITLE}")"
        BRANCH_NAME="${BLOB_ID}"

        if committed_changes "${BRANCH_NAME}" "${COMMIT_MESSAGE}" "${COMMIT_USERNAME}" "${COMMIT_USEREMAIL}" "${GH_USER}" "${GH_TOKEN}";
        then
            create_pr "${BRANCH_NAME}" "${PR_BASE}" "${PR_TITLE}" "${PR_MESSAGE}" "${PR_LABELS}" "${GH_USER}" "${GH_TOKEN}"
        else
            echo ""
        fi

        #git stash && git checkout "${COMMIT_SAVEPOINT}"
    fi
done
