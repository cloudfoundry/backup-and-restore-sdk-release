#!/usr/bin/env bash
set -euo pipefail

# Prior to postgres 10, x.y was a major version (i.e. 9.6).
# As of postgres 10, they follow semver.
MAJOR=${MAJOR:?Required: postgres major version}

MINOR=${MINOR:?Required: postgres minor version}

sdk_root="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null 2>&1 && pwd )"
old_blob="$(
  if ! grep --only-matching "postgres/postgresql-${MAJOR}\.\d\+.tar.gz" "${sdk_root}/config/blobs.yml"; then
      echo "could not find MAJOR version v${MAJOR} in blobs.yml" 1>&2
  fi
)"
new_version="${MAJOR}.${MINOR}"

sed -i "" \
    "s/postgresql-.*.tar.gz/postgresql-${new_version}.tar.gz/" \
    "${sdk_root}/packages/database-backup-restorer-postgres-${MAJOR}/spec"

sed -i "" \
    "s/POSTGRES_VERSION=.*$/POSTGRES_VERSION=${new_version}/" \
    "${sdk_root}/packages/database-backup-restorer-postgres-${MAJOR}/packaging"

wget "--directory-prefix=${sdk_root}" \
  "https://ftp.postgresql.org/pub/source/v${new_version}/postgresql-${new_version}.tar.gz"

bosh remove-blob "--dir=${sdk_root}" "$old_blob"

bosh add-blob "--dir=${sdk_root}" \
  "${sdk_root}/postgresql-${new_version}.tar.gz" \
  "postgres/postgresql-${new_version}.tar.gz"

rm "postgresql-${new_version}.tar.gz"

bosh upload-blobs "--dir=${sdk_root}"

