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
: "${DOWNLOADED_FILENAME:?}"
: "${ALL_VERSIONS:?}"

source "${SCRIPT_DIR}/functions.sh"

REPO_ROOT="$(git -C "$(realpath "$(dirname "${AUTOBUMP_DESCRIPTOR}")")" rev-parse --show-toplevel)"

COMMIT_SAVEPOINT="$(git -C "${REPO_ROOT}" rev-parse HEAD)"
for BLOB_ID in $(current_blobs_names "${BLOBS_PREFIX}");
do
    setup_private_blobstore_config "${AWS_ACCESS_KEY_ID}" "${AWS_SECRET_ACCESS_KEY}"

    if callback_defined "extract_version_callback";
    then PREV_VERSION="$(extract_version_callback "${BLOB_ID}")"
    else PREV_VERSION="$(echo "${BLOB_ID}" | grep -Eo '[0-9]+(\.[0-9]+)*')"
    fi

    if ! callback_defined "new_version_callback";
    then
        NEW_VERSION="$(pick_cadidate_version "${PREV_VERSION}" "${ALL_VERSIONS}" "AUTO")"
    else
        BUMP_STRATEGY="$(new_version_callback "${PREV_VERSION}" "${ALL_VERSIONS}")"

        if [[ "${BUMP_STRATEGY}" == "MAJOR" ]] || [[ "${BUMP_STRATEGY}" == "MINOR" ]] || [[ "${BUMP_STRATEGY}" == "PATCH" ]] ;
        then NEW_VERSION="$(pick_cadidate_version "${PREV_VERSION}" "${ALL_VERSIONS}" "${BUMP_STRATEGY}")"
        else NEW_VERSION="${BUMP_STRATEGY}" # new_version_callback can return one of the allowed strategies or a specific version
        fi
    fi

    DOWNLOAD_URL="$(download_url_callback "${NEW_VERSION}")"

    if callback_defined "new_blobid_callback";
    then NEW_BLOB_ID="$(new_blobid_callback "${BLOB_ID}" "${PREV_VERSION}" "${NEW_VERSION}" )"
    else NEW_BLOB_ID="$(echo "${BLOB_ID}" | grep "${PREV_VERSION}" | sed "s/${PREV_VERSION}/${NEW_VERSION}/")"
    fi

    if blobs_are_equal "${BLOB_ID}" "${NEW_BLOB_ID}" "${PREV_VERSION}" "${NEW_VERSION}";
    then
        echo "${BLOB_ID} is up-to-date"
    else
        NEW_TARFILE="$(download_version "${NEW_VERSION}" "${DOWNLOAD_URL}" "${DOWNLOADED_FILENAME}")"

        if callback_defined "checksum_callback";
        then # Callback for verifying checksum is defined in AUTOBUMP_DESCRIPTOR script. Let's call it!"
            checksum_callback "${NEW_VERSION}" "${NEW_TARFILE}"
        fi

        replace_blob "${BLOB_ID}" "${NEW_BLOB_ID}" "${PREV_VERSION}" "${NEW_VERSION}" "${NEW_TARFILE}" "${BLOBS_PREFIX}"

        # Following function **SHOULD NEVER BE PERFORMED BEFORE** the replace_blob function
        pushd "${REPO_ROOT}" >/dev/null
        if callback_defined "replace_references_callback";
        then replace_references_callback "${BLOB_ID}" "${NEW_BLOB_ID}" "${PREV_VERSION}" "${NEW_VERSION}" "${NEW_TARFILE}" "${BLOBS_PREFIX}"
        else replace_references "${BLOB_ID}" "${NEW_BLOB_ID}" "${PREV_VERSION}" "${NEW_VERSION}" "${NEW_TARFILE}" "${BLOBS_PREFIX}"
        fi
        popd >/dev/null

        rm -f "${NEW_TARFILE}"

        EXPANDED_COMMIT_MESSAGE="$(safely_expand_variables "${COMMIT_MESSAGE}")"
        EXPANDED_PR_MESSAGE="$(safely_expand_variables "${PR_MESSAGE}")"
        EXPANDED_PR_TITLE="$(safely_expand_variables "${PR_TITLE}")"
        BRANCH_NAME="${BLOB_ID}"

        if committed_changes "${BRANCH_NAME}" "${EXPANDED_COMMIT_MESSAGE}" "${COMMIT_USERNAME}" "${COMMIT_USEREMAIL}" "${GH_USER}" "${GH_TOKEN}";
        then
            create_pr "${BRANCH_NAME}" "${PR_BASE}" "${EXPANDED_PR_TITLE}" "${EXPANDED_PR_MESSAGE}" "${PR_LABELS}" "${GH_USER}" "${GH_TOKEN}"
        fi

        git -C "${REPO_ROOT}" stash
        git -C "${REPO_ROOT}" checkout "${COMMIT_SAVEPOINT}"
    fi
done
