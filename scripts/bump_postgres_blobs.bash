#!/usr/bin/env bash
set -euo pipefail

sdk_root="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null 2>&1 && pwd )"
old_blob="$( grep --only-matching "postgres/postgresql-${MAJOR}\.\d\+.tar.gz" "${sdk_root}/config/blobs.yml" )"
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

