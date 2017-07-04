# backup-and-restore-sdk BOSH release

A backup and restore sdk for BOSH releases for e.g. backing up a number of different databases.

### Usage for database backup

Co-locate the `database-backuper` job on the database VM that should be backed up:

```yaml
---
instance_groups:
- name: postgres
  jobs:
  - name: postgres-server
    release: postgres-release
  - name: database-backuper                       <<
    release: backup-and-restore-sdk-release  <<
...
```

Template a `config.json` as follows:

```json
{
  "username": "db user",
  "password": "db password",
  "host": "db host",
  "port": "db port",
  "adapter": "db adapter; see 'Supported database adapters'",
  "database": "name of database to back up"
}
```

In your release backup script, call database-backuper/bin/backup:

```bash
/var/vcap/jobs/database-backuper/bin/backup --config /path/to/config.json --artifact-file /path/to/artifact/file
```

If using BOSH Backup and Restore, ensure that your artifact file lives inside the $BBR_ARTIFACT_DIRECTORY. For example:

```bash
/var/vcap/jobs/database-backuper/bin/backup --config /path/to/config.json --artifact-file $BBR_ARTIFACT_DIRECTORY/sqlDump
```

### Supported database adapters
* `postgres` - packages 9.4.11
