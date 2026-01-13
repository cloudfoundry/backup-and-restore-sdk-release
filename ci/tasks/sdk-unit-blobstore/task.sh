#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

cd "backup-and-restore-sdk-release/src/${PACKAGE_NAME}"
# shellcheck disable=SC2086
go run github.com/onsi/ginkgo/v2/ginkgo run -r --keep-going ${GINKGO_EXTRA_FLAGS:-}
