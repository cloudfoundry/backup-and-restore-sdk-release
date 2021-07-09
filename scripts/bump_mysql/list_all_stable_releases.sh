#!/usr/bin/env bash
set -euo pipefail

HTML="$(curl -s -L https://downloads.mysql.com/archives/community/)"
VALUES="$(echo "${HTML}" | xmllint --html --xpath "//select[@id='version']/option[@value=text()]/@value" - 2>/dev/null)"
VERSIONS="$(echo "${VALUES}" | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+[a-zA-Z]?')"

echo "${VERSIONS}" | sort -t "." -k1,1n -k2,2n -k3,3n
