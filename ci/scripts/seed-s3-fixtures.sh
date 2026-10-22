#!/usr/bin/env bash
#
# Seeds the pre-existing S3 fixtures the s3-blobstore-backup-restore contract
# tests (src/s3-blobstore-backup-restore/s3bucket) expect, in the AWS account
# 504405067906.
#
#   ci-bbr-sdk-large-blob              (versioned)   big_file, copied by
#                                                    CopyVersion with an
#                                                    explicit version ID
#   ci-bbr-sdk-large-blob-unversioned  (unversioned) big_file, copied by
#                                                    CopyBlobFromBucket
#   ci-bbr-sdk-big-bucket              (versioned)   exactly 2001 object versions
#   ci-bbr-sdk-unversioned-big-bucket  (unversioned) exactly 2001 objects
#   ci-bbr-sdk-large-number-of-files   (unversioned) exactly 2001 objects; the
#                                                    s3 system test backs it
#                                                    up and expects > 1900
#
# big_file must be larger than 1024 MiB: above that, the SDK copies with
# multipart (100 MiB parts) instead of a single CopyObject, and that path is
# what these tests exercise. 2001 objects is just over two 1000-key pages,
# so listing has to paginate.
#
# At the end the script prints the values the pipeline passes to the tests
# (bucket names, region, version ID and sha1s).
#
# Safe to re-run: fixtures that are already in place are left alone and
# their values are reported again. Partly seeded buckets are completed, and
# extra object versions are pruned.
#
# Needs: aws CLI v2, ~BIG_FILE_SIZE_MB of free space in TMPDIR, and time to
# upload big_file twice.
#
# Usage:
#   ./ci/scripts/seed-s3-fixtures.sh
#   DRY_RUN=true ./ci/scripts/seed-s3-fixtures.sh       # report only, change nothing
#   BIG_FILE_SIZE_MB=2048 ./ci/scripts/seed-s3-fixtures.sh

set -Eeuo pipefail

[ -z "${DEBUG:-}" ] || set -x

trap 'echo "ERROR: command failed (exit $?) at line ${LINENO}: ${BASH_COMMAND}" >&2' ERR

EXPECTED_ACCOUNT_ID="504405067906"
REGION="eu-north-1"
DRY_RUN="${DRY_RUN:-false}"

BIG_FILE_KEY="big_file"
BIG_FILE_SIZE_MB="${BIG_FILE_SIZE_MB:-1536}"
BIG_FILE_VERSIONED_BUCKET="ci-bbr-sdk-large-blob"
BIG_FILE_UNVERSIONED_BUCKET="ci-bbr-sdk-large-blob-unversioned"

MANY_OBJECTS_COUNT=2001
MANY_VERSIONS_BUCKET="ci-bbr-sdk-big-bucket"
MANY_OBJECTS_BUCKET="ci-bbr-sdk-unversioned-big-bucket"
LARGE_NUMBER_OF_FILES_BUCKET="ci-bbr-sdk-large-number-of-files"

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "${WORK_DIR}"' EXIT

check_account() {
  local account_id
  account_id="$(aws sts get-caller-identity --query Account --output text)"
  if [[ "${account_id}" != "${EXPECTED_ACCOUNT_ID}" ]]; then
    echo "ERROR: logged in to AWS account ${account_id}, expected ${EXPECTED_ACCOUNT_ID}" >&2
    exit 1
  fi
}

check_bucket() {
  local bucket="$1" want_versioning="$2"
  local versioning
  versioning="$(aws s3api get-bucket-versioning --bucket "${bucket}" --region "${REGION}" \
    --query Status --output text)"
  if [[ "${want_versioning}" == "true" && "${versioning}" != "Enabled" ]]; then
    echo "ERROR: ${bucket} must be versioned (versioning is '${versioning}')" >&2
    exit 1
  fi
  if [[ "${want_versioning}" == "false" && "${versioning}" != "None" ]]; then
    echo "ERROR: ${bucket} must never have been versioned (versioning is '${versioning}')" >&2
    exit 1
  fi
}

# --- big_file -----------------------------------------------------------

BIG_FILE_PATH=""
BIG_FILE_SHA1=""

generate_big_file() {
  if [[ -n "${BIG_FILE_PATH}" ]]; then
    return
  fi
  BIG_FILE_PATH="${WORK_DIR}/${BIG_FILE_KEY}"
  echo "generating ${BIG_FILE_SIZE_MB} MiB ${BIG_FILE_KEY}..."
  head -c "$(( BIG_FILE_SIZE_MB * 1024 * 1024 ))" /dev/urandom > "${BIG_FILE_PATH}"
  if command -v shasum >/dev/null; then
    BIG_FILE_SHA1="$(shasum -a 1 "${BIG_FILE_PATH}" | awk '{print $1}')"
  else
    BIG_FILE_SHA1="$(sha1sum "${BIG_FILE_PATH}" | awk '{print $1}')"
  fi
}

# Prints "<sha1> <version-id>" for big_file in the bucket, or nothing if the
# object isn't there. The sha1 comes from the object's metadata, which this
# script sets on upload, so it doesn't need re-downloading.
big_file_state() {
  local bucket="$1"
  aws s3api head-object --bucket "${bucket}" --key "${BIG_FILE_KEY}" --region "${REGION}" \
    --query '[Metadata.sha1, VersionId]' --output text 2>/dev/null || true
}

seed_big_file() {
  local bucket="$1" versioned="$2"
  check_bucket "${bucket}" "${versioned}"

  local state
  state="$(big_file_state "${bucket}")"
  if [[ -n "${state}" ]]; then
    local sha1
    sha1="$(awk '{print $1}' <<< "${state}")"
    if [[ -z "${sha1}" || "${sha1}" == "None" ]]; then
      echo "ERROR: s3://${bucket}/${BIG_FILE_KEY} exists but has no sha1 metadata." >&2
      echo "       It wasn't uploaded by this script; delete it and re-run." >&2
      exit 1
    fi
    echo "exists:  s3://${bucket}/${BIG_FILE_KEY}"
    return
  fi

  echo "upload:  s3://${bucket}/${BIG_FILE_KEY} (${BIG_FILE_SIZE_MB} MiB)"
  if [[ "${DRY_RUN}" == "true" ]]; then
    return
  fi
  generate_big_file
  aws s3 cp "${BIG_FILE_PATH}" "s3://${bucket}/${BIG_FILE_KEY}" \
    --region "${REGION}" \
    --metadata "sha1=${BIG_FILE_SHA1}" \
    --only-show-errors
}

# --- many objects -------------------------------------------------------

# The AWS CLI applies --query to each page of a paginated listing, so each
# function below sums or concatenates the per-page results.
# The backticks are JMESPath literals, not shell expansions.
# shellcheck disable=SC2016
count_objects() {
  local bucket="$1" versioned="$2"
  if [[ "${versioned}" == "true" ]]; then
    aws s3api list-object-versions --bucket "${bucket}" --region "${REGION}" \
      --query 'length(Versions || `[]`)' --output text
  else
    aws s3api list-objects-v2 --bucket "${bucket}" --region "${REGION}" \
      --query 'length(Contents || `[]`)' --output text
  fi | awk '{ total += $1 } END { print total + 0 }'
}

# Deletes every non-latest version, so each key has exactly one version.
# Extra versions appear when an upload is retried. Deletes in batches of up
# to 1000 versions per request (the delete-objects limit).
# shellcheck disable=SC2016
prune_old_versions() {
  local bucket="$1"
  local list_file batch_file
  list_file="${WORK_DIR}/${bucket}.old-versions"
  aws s3api list-object-versions --bucket "${bucket}" --region "${REGION}" \
    --query 'Versions[?IsLatest==`false`].[Key, VersionId]' --output text \
    | awk 'NF == 2 && $1 != "None"' > "${list_file}"

  local total
  total="$(wc -l < "${list_file}" | tr -d ' ')"
  if (( total == 0 )); then
    return
  fi

  echo "prune:   ${total} old version(s) from s3://${bucket}"
  batch_file="${WORK_DIR}/${bucket}.batch.json"
  local offset=0
  while (( offset < total )); do
    # One awk selects the batch's lines; piping through head here would
    # SIGPIPE the writer and, with pipefail, silently kill the script.
    awk -v first="$(( offset + 1 ))" -v last="$(( offset + 1000 ))" '
      BEGIN { printf "{\"Objects\":["; n = 0 }
      NR >= first && NR <= last {
        printf "%s{\"Key\":\"%s\",\"VersionId\":\"%s\"}", (n++ ? "," : ""), $1, $2
      }
      END { printf "],\"Quiet\":true}" }' "${list_file}" > "${batch_file}"
    # With Quiet, per-object failures come back in Errors and the exit
    # status is still 0, so check for them.
    local errors
    errors="$(aws s3api delete-objects --bucket "${bucket}" --region "${REGION}" \
      --delete "file://${batch_file}" --query 'Errors[].[Key, Code, Message]' --output text)"
    if [[ -n "${errors}" && "${errors}" != "None" ]]; then
      echo "ERROR: some old versions in s3://${bucket} could not be deleted:" >&2
      echo "${errors}" | head -n 5 >&2
      exit 1
    fi
    offset=$(( offset + 1000 ))
  done
}

# Prints the keys in the bucket, one per line.
list_keys() {
  local bucket="$1"
  aws s3api list-objects-v2 --bucket "${bucket}" --region "${REGION}" \
    --query 'Contents[].Key' --output text \
    | tr '\t' '\n' | awk 'NF && $0 != "None"'
}

seed_many_objects() {
  local bucket="$1" versioned="$2"
  check_bucket "${bucket}" "${versioned}"

  if [[ "${versioned}" == "true" && "${DRY_RUN}" != "true" ]]; then
    prune_old_versions "${bucket}"
  fi

  local count
  count="$(count_objects "${bucket}" "${versioned}")"
  if [[ "${count}" == "${MANY_OBJECTS_COUNT}" ]]; then
    echo "exists:  s3://${bucket} (${count} objects)"
    return
  fi

  # Not seeded, or partly seeded: upload only the missing keys, so existing
  # keys in a versioned bucket don't get a second version.
  local dir
  dir="${WORK_DIR}/${bucket}"
  mkdir -p "${dir}"
  local existing="${dir}.existing"
  list_keys "${bucket}" | sort > "${existing}"

  local i key missing=0
  for i in $(seq -f '%04g' 1 "${MANY_OBJECTS_COUNT}"); do
    key="blob-${i}"
    if ! grep -qx "${key}" "${existing}"; then
      echo "${i}" > "${dir}/${key}"
      missing=$(( missing + 1 ))
    fi
  done

  echo "upload:  s3://${bucket} (${missing} missing of ${MANY_OBJECTS_COUNT} objects)"
  if [[ "${DRY_RUN}" == "true" ]]; then
    return
  fi
  if (( missing > 0 )); then
    aws s3 cp "${dir}" "s3://${bucket}/" --recursive --region "${REGION}" --only-show-errors
  fi

  if [[ "${versioned}" == "true" ]]; then
    prune_old_versions "${bucket}"
  fi

  count="$(count_objects "${bucket}" "${versioned}")"
  if [[ "${count}" != "${MANY_OBJECTS_COUNT}" ]]; then
    echo "ERROR: s3://${bucket} holds ${count} objects after upload, expected ${MANY_OBJECTS_COUNT}." >&2
    echo "       It may contain objects this script didn't create; check and re-run." >&2
    exit 1
  fi
}

# --- report -------------------------------------------------------------

report() {
  if [[ "${DRY_RUN}" == "true" ]]; then
    return
  fi

  local versioned_state unversioned_state
  versioned_state="$(big_file_state "${BIG_FILE_VERSIONED_BUCKET}")"
  unversioned_state="$(big_file_state "${BIG_FILE_UNVERSIONED_BUCKET}")"

  local versioned_sha1 versioned_version_id unversioned_sha1
  read -r versioned_sha1 versioned_version_id <<< "${versioned_state}"
  read -r unversioned_sha1 _ <<< "${unversioned_state}"

  cat <<EOF

Fixture values for the pipeline (contract-tests, aws-s3-blobstore-contract-tests):

  region:                          ${REGION}
  versioned big_file bucket:       ${BIG_FILE_VERSIONED_BUCKET}
  versioned big_file version ID:   ${versioned_version_id}
  versioned big_file sha1:         ${versioned_sha1}
  unversioned big_file bucket:     ${BIG_FILE_UNVERSIONED_BUCKET}   (S3_BIG_FILE_BUCKET)
  unversioned big_file sha1:       ${unversioned_sha1}
  2001-version bucket:             ${MANY_VERSIONS_BUCKET}
  2001-object bucket:              ${MANY_OBJECTS_BUCKET}
EOF
}

main() {
  if (( BIG_FILE_SIZE_MB <= 1024 )); then
    echo "ERROR: BIG_FILE_SIZE_MB must be over 1024, so the SDK copies it with multipart" >&2
    exit 1
  fi

  check_account

  seed_big_file "${BIG_FILE_VERSIONED_BUCKET}" true
  seed_big_file "${BIG_FILE_UNVERSIONED_BUCKET}" false
  seed_many_objects "${MANY_VERSIONS_BUCKET}" true
  seed_many_objects "${MANY_OBJECTS_BUCKET}" false
  seed_many_objects "${LARGE_NUMBER_OF_FILES_BUCKET}" false

  report
}

main "$@"
