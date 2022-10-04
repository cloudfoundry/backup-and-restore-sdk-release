#!/usr/bin/env bash

# For some reason these variables are always mandatory even if
# they are empty they need to exists for the tests to succeed
set -euo pipefail

export PG_DUMP_9_4_PATH=""
export PG_RESTORE_9_4_PATH=""

export PG_DUMP_9_6_PATH=""
export PG_RESTORE_9_6_PATH=""

export PG_DUMP_10_PATH=""
export PG_RESTORE_10_PATH=""

export PG_DUMP_11_PATH=""
export PG_RESTORE_11_PATH=""

export PG_DUMP_13_PATH=""
export PG_RESTORE_13_PATH=""

export PG_CLIENT_PATH=""

export MARIADB_DUMP_PATH=""
export MARIADB_CLIENT_PATH=""

export MYSQL_DUMP_5_6_PATH=""
export MYSQL_CLIENT_5_6_PATH=""

export MYSQL_DUMP_5_7_PATH=""
export MYSQL_CLIENT_5_7_PATH=""

export MYSQL_DUMP_8_0_PATH=""
export MYSQL_CLIENT_8_0_PATH=""

