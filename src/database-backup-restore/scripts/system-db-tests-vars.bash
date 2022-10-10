#!/usr/bin/env bash

# For some reason these variables are always mandatory even if
# they are empty they need to exists for the tests to succeed
set -euo pipefail

export PG_DUMP_9_4_PATH="/var/vcap/packages/database-backup-restorer-postgres-9.4/bin/pg_dump"
export PG_RESTORE_9_4_PATH="/var/vcap/packages/database-backup-restorer-postgres-9.4/bin/pg_restore"

export PG_DUMP_9_6_PATH="/var/vcap/packages/database-backup-restorer-postgres-9.6/bin/pg_dump"
export PG_RESTORE_9_6_PATH="/var/vcap/packages/database-backup-restorer-postgres-9.6/bin/pg_restore"

export PG_DUMP_10_PATH="/var/vcap/packages/database-backup-restorer-postgres-10/bin/pg_dump"
export PG_RESTORE_10_PATH="/var/vcap/packages/database-backup-restorer-postgres-10/bin/pg_restore"

export PG_DUMP_11_PATH="/var/vcap/packages/database-backup-restorer-postgres-11/bin/pg_dump"
export PG_RESTORE_11_PATH="/var/vcap/packages/database-backup-restorer-postgres-11/bin/pg_restore"

export PG_DUMP_13_PATH="/var/vcap/packages/database-backup-restorer-postgres-13/bin/pg_dump"
export PG_RESTORE_13_PATH="/var/vcap/packages/database-backup-restorer-postgres-13/bin/pg_restore"

export PG_CLIENT_PATH="/var/vcap/packages/database-backup-restorer-postgres-13/bin/psql"

export MARIADB_DUMP_PATH="/var/vcap/packages/database-backup-restorer-mariadb/bin/mysqldump"
export MARIADB_CLIENT_PATH="/var/vcap/packages/database-backup-restorer-mariadb/bin/mysql"

export MYSQL_DUMP_5_6_PATH="/var/vcap/packages/database-backup-restorer-mysql-5.6/bin/mysqldump"
export MYSQL_CLIENT_5_6_PATH="/var/vcap/packages/database-backup-restorer-mysql-5.6/bin/mysql"

export MYSQL_DUMP_5_7_PATH="/var/vcap/packages/database-backup-restorer-mysql-5.7/bin/mysqldump"
export MYSQL_CLIENT_5_7_PATH="/var/vcap/packages/database-backup-restorer-mysql-5.7/bin/mysql"

export MYSQL_DUMP_8_0_PATH="/var/vcap/packages/database-backup-restorer-mysql-8.0/bin/mysqldump"
export MYSQL_CLIENT_8_0_PATH="/var/vcap/packages/database-backup-restorer-mysql-8.0/bin/mysql"

