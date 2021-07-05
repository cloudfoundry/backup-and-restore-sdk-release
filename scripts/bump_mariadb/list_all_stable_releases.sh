#!/usr/bin/env bash
set -euo pipefail

HTML="$(curl -s -L https://downloads.mariadb.org/mariadb/+releases/)"
HREFS="$(echo "${HTML}" | xmllint --html --xpath "//table[@id='download']/tbody/tr[td[3]='Stable']/td[1]/a/@href" - 2>/dev/null)"
VERSIONS="$(echo "${HREFS}" | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+')"

echo "${VERSIONS}" | sort -t "." -k1,1n -k2,2n -k3,3n
