#!/usr/bin/env bash
set -euo pipefail

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
