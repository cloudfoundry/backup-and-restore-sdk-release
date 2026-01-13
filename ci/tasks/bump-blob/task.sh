#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

shopt -s extglob

cp -r bosh-release/. bosh-release-updated

NEW_VERSION="$(cat "${VERSION_FILE:?}")"

pushd bosh-release-updated
  if [ -n "$PRIVATE_YML" ]
  then
    echo "$PRIVATE_YML" > "config/private.yml"
  elif [ -n "$GCP_KEY" ]
  then
    cat <<EOF > "config/private.yml"
  blobstore:
    options:
      credentials_source: static
      json_key: '${GCP_KEY}'
EOF
  else
    echo "---
  blobstore:
    provider: s3
    options:
      access_key_id: $AWS_ACCESS_KEY_ID
      secret_access_key: $AWS_SECRET_ACCESS_KEY" > config/private.yml
    if [[ -n "$AWS_ROLE_ARN" ]]; then
      echo "      assume_role_arn: $AWS_ROLE_ARN" >> config/private.yml
    fi
  fi

  set -eux

  if git show-ref --quiet "refs/heads/$GIT_BRANCH"; then
    git checkout "$GIT_BRANCH"
  else
    git checkout -b "$GIT_BRANCH"
  fi

  BLOB_NAME_WITH_PATH="$BLOB_NAME"
  if [[ -n "$BLOB_DIRECTORY" ]]; then
    BLOB_NAME_WITH_PATH="${BLOB_DIRECTORY%/}/${BLOB_NAME}"
  fi

  # ${KEEP_BLOBS_FILTER/,/\\\|}
  # KEEP_BLOBS_FILTER expects a csv and will transform the values into a grep "or" matcher by replacing `,` with `\|`
  # This grep -v needs to happen before we pipe the output into xargs, because that will also trim newlines.

  # Some blobs in some releases need to be available in two versions to keep xenial compatibility. e.g. talloc in smb-vol-release.
  # bbr-sdk is another example, it has postgresql-11, postgresql-13, postgresql-15
  # in these releases, current filename could contain a value similar to: 'talloc-2.4.0.tar.gz xenial-talloc-2.3.1.tar.gz' (actual value from CI output)

  # this is problematic for two reasons.
  # a)
  # running "bosh remove-blob 'talloc-2.4.0.tar.gz xenial-talloc-2.3.1.tar.gz'" will exit 0 but not remove anything.
  # b)
  # it could accidentially remove a blob that we actually still need.

 if [[ -n ${KEEP_BLOBS_FILTER} ]]; then
    CURRENT_FILENAME=$(bosh blobs --column=path | grep "$BLOB_NAME_WITH_PATH"| grep -v "${KEEP_BLOBS_FILTER/,/\\|}" | xargs) # xargs is used to trim the trailing spaces
 else
    CURRENT_FILENAME=$(bosh blobs --column=path | grep "$BLOB_NAME_WITH_PATH"| xargs)
 fi
  # On 2023-11-30, ksemenov tested the version-getting regex as follows:
  # $ for f in autoconf-2.71.tar.gz automake-1.16.5.tar.xz libtool-2.4.7.tar.xz json-c/json-c-0.17-nodoc.tar.gz jq/jq-1.7-linux-amd64; do
  #     echo $f | sed -r 's/^.+-([0-9.]*[0-9]).*$/\1/'
  #   done
  CURRENT_VERSION=$(echo "$CURRENT_FILENAME" | sed -r 's/^.+-([0-9.]*[0-9]).*$/\1/')


  # On 2023-09-12, gds tested the version-getting regex as follows:
  # $ BLOB_EXTENSION="tar.gz" echo autoconf-2.71.tar.gz | sed -r 's/^.+-([0-9.]*[0-9])\.?'"$BLOB_EXTENSION"'.*/\1/'
  # 2.71
  # 1.16.5
  # 2.4.7
  # 0.17
  # 1.7

  if [[ "${CURRENT_VERSION}" == "${NEW_VERSION}" ]]; then
    echo "Versions already match. Skipping update."
    exit 0
  fi

  NEW_FILENAME=$(ls ../distributed-package/${BLOB_NAME%-}?(-)${NEW_VERSION}[-.]${BLOB_EXTENSION#[-.]})
  # On 2023-11-21, ksemenov tested the file-matching pattern as follows:
  # Given a directory with the following files:
  #    autoconf-2.71-linux-x86.tar.gz	json-c-0.17-nodoc.tar.gz	openssl-1.2.4.tar.gz
  #    autoconf-2.71.tar.gz						libtool-2.4.7.tar.xz      tcl8.6.13.tar.gz
  ##
  #  $ BLOB_NAME="autoconf-"; NEW_VERSION="2.71"; BLOB_EXTENSION="tar.gz"; ls "${BLOB_NAME%-}-${NEW_VERSION}"[-.]"${BLOB_EXTENSION#[-.]}"
  #  autoconf-2.71.tar.gz
  #  $ BLOB_NAME="autoconf-"; NEW_VERSION="2.71"; BLOB_EXTENSION="linux-x86.tar.gz"; ls "${BLOB_NAME%-}-${NEW_VERSION}"[-.]"${BLOB_EXTENSION#[-.]}"
  #  autoconf-2.71-linux-x86.tar.gz
  #  $ BLOB_NAME="json-c"; NEW_VERSION="0.17"; BLOB_EXTENSION="-nodoc.tar.gz"; ls "${BLOB_NAME%-}-${NEW_VERSION}"[-.]"${BLOB_EXTENSION#[-.]}"
  #  json-c-0.17-nodoc.tar.gz
  #  $ BLOB_NAME="json-c"; NEW_VERSION="0.17"; BLOB_EXTENSION="nodoc.tar.gz"; ls "${BLOB_NAME%-}-${NEW_VERSION}"[-.]"${BLOB_EXTENSION#[-.]}"
  #  json-c-0.17-nodoc.tar.gz
  #  $ BLOB_NAME="json-c-"; NEW_VERSION="0.17"; BLOB_EXTENSION="nodoc.tar.gz"; ls "${BLOB_NAME%-}-${NEW_VERSION}"[-.]"${BLOB_EXTENSION#[-.]}"
  #  json-c-0.17-nodoc.tar.gz
  #  $ BLOB_NAME="libtool"; NEW_VERSION="2.4.7"; BLOB_EXTENSION=".tar.xz"; ls "${BLOB_NAME%-}-${NEW_VERSION}"[-.]"${BLOB_EXTENSION#[-.]}"
  #  libtool-2.4.7.tar.xz
  #  $ BLOB_NAME="openssl"; NEW_VERSION="1.2.4"; BLOB_EXTENSION="tar.gz"; ls "${BLOB_NAME%-}-${NEW_VERSION}"[-.]"${BLOB_EXTENSION#[-.]}"
  #  openssl-1.2.4.tar.gz
  #  $ BLOB_NAME="tcl"; NEW_VERSION="8.6.13"; BLOB_EXTENSION="-src.tar.gz"; ls "${BLOB_NAME%-}-${NEW_VERSION}"[-.]"${BLOB_EXTENSION#[-.]}"
  #  tcl8.6.13.tar.gz
  #
  #  Below there's a code snippet that you could stick in a file to test new additions:
  #
  # [ -z "${DEBUG:-}" ] || set -x
  # set -e
  # shopt -s extglob
  # cd $( mktemp -d ) || exit 1
  # touch \
  #   autoconf-2.71-linux-x86.tar.gz \
  #   json-c-0.17-nodoc.tar.gz \
  #   openssl-1.2.4.tar.gz \
  #   autoconf-2.71.tar.gz \
  #   libtool-2.4.7.tar.xz \
  #   tcl8.6.13-src.tar.gz \
  #
  # function check_file_exists() {
  #   local BLOB_NAME="${1:?}"
  #   local NEW_VERSION="${2:?}"
  #   local BLOB_EXTENSION="${3:?}"
  #
  #   ls "${BLOB_NAME%-}"?(-)"${NEW_VERSION}"[-.]"${BLOB_EXTENSION#[-.]}" || exit 1
  # }
  #
  # check_file_exists "autoconf-" "2.71" "tar.gz"
  # check_file_exists "autoconf-" "2.71" "linux-x86.tar.gz"
  # check_file_exists "json-c" "0.17" "-nodoc.tar.gz"
  # check_file_exists "json-c" "0.17" "nodoc.tar.gz"
  # check_file_exists "json-c-" "0.17" "nodoc.tar.gz"
  # check_file_exists "libtool" "2.4.7" ".tar.xz"
  # check_file_exists "openssl" "1.2.4" "tar.gz"
  # check_file_exists "tcl" "8.6.13" "-src.tar.gz"

  if [[ -z ${BLOB_DIRECTORY} ]]; then
    NEW_BLOBNAME="${NEW_FILENAME#../distributed-package/}"
  else
    NEW_BLOBNAME="${BLOB_DIRECTORY%/}/${NEW_FILENAME#../distributed-package/}"
  fi

  echo "removing old blob(s) $CURRENT_FILENAME"
  # check if we have " "(spaces) in CURRENT_FILENAME. It would indicate that we are about to remove more than one blob.
  if [[ $(tr -dc ' ' <<<"$CURRENT_FILENAME" | wc -c) -gt 0 ]]; then
    if [[ $REMOVE_MULTIPLE_BLOBS == true ]]; then
      echo "${CURRENT_FILENAME}" | xargs -n1 bosh remove-blob
    else
      echo -e "It seems this would remove more than one blob. But remove multiple blobs is set to: $REMOVE_MULTIPLE_BLOBS\n bailing out to avoid removing the wrong blob"
      exit 1
    fi
  else
    bosh remove-blob "${CURRENT_FILENAME}"
  fi

  echo "updating ${BLOB_NAME} blob from ${CURRENT_VERSION} to ${NEW_VERSION}"
  bosh add-blob "../distributed-package/${NEW_FILENAME}" "$NEW_BLOBNAME"

  # check if the blob we just added was already present. If the size and the checksum didn't change, we just added the same blob because the bumper was rerun for any reason. In that case the diff look similar to:
  # git diff -U0 config/blobs.yml
  # diff --git a/config/blobs.yml b/config/blobs.yml
  # index 601299f..79fed08 100644
  # --- a/config/blobs.yml
  # +++ b/config/blobs.yml
  # @@ -3 +2,0 @@ clamav-windows/clamav-1.0.5-win-x64-portable.zip:
  # -  object_id: fd44c156-4f74-41e7-7f5f-bc3b0193fea9
  #
  # we ignore the first 5 lines via tail, because that's just metadata about the change..
  # a real change to the file, would look like this:
  #
  # -  size: 16675830
  # -  object_id: fd44c156-4f74-41e7-7f5f-bc3b0193fea9
  # -  sha: sha256:2e925af686e7b97e8ff17ba4ec7895fe1fafbd7c663ad00f156cd96bb4ffe7cf
  # +  size: 603
  # +  sha: sha256:a53b730d7852692fea0a5e97a6458aa88783f39333e78ec844adcf23c8fcd358

  if [[ $(git diff -U0 config/blobs.yml | tail -n +6 | wc -l) -eq 1 ]]; then
    echo -e "skipping commit. The resulting diff: \n$(git diff config/blobs.yml)\nseems to bump to an identical blob."
    exit 0
  fi

  bosh upload-blobs
  git add config/blobs.yml

  if [[ -n "${UPDATE_REFERENCES:-""}" ]]
  then
    if [[ "$UPDATE_REFERENCES" =~ ^only-variable-name: ]]; then
      VARIABLE_NAME="${UPDATE_REFERENCES#only-variable-name:}"
      git grep -E -l --threads=1 "(^|[^A-Z0-9_])${VARIABLE_NAME}.?=${CURRENT_VERSION}" -- jobs/ packages/ | xargs sed -re "s/(^|[^A-Z0-9_])(${VARIABLE_NAME}.?=)${CURRENT_VERSION}/\\1\\2${NEW_VERSION}/g" -i'~'
    else
      git grep -l --threads=1 "${CURRENT_VERSION}" -- jobs/ packages/ | xargs sed -e "s/${CURRENT_VERSION}/${NEW_VERSION}/g" -i'~'
      mv packages/"$BLOB_NAME"{"${CURRENT_VERSION}","${NEW_VERSION}"}
    fi
    git add packages/ jobs/
  fi


  git config user.name "${GIT_USERNAME}"
  git config user.email "${GIT_EMAIL}"
  git commit -m "Bump ${BLOB_NAME} from ${CURRENT_VERSION} to ${NEW_VERSION}"
popd
