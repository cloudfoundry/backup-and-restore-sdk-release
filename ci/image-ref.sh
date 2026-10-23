#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
ref="$1:$2"
digest=$(awk -v ref="$ref" '$1 == ref { print $2 }' "$script_dir/image-pins.txt")
if [[ ! "$digest" =~ ^sha256:[0-9a-f]{64}$ ]]; then
  echo "No reviewed image digest for $ref; update ci/image-pins.txt before building." >&2
  exit 1
fi
printf '%s@%s\n' "$ref" "$digest"
