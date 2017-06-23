# database-backup-and-restore BOSH release

A BOSH release for backing up a number of different databases.

### Usage

Co-locate the `database-backuper` job on the database VM that should be backed up:

```yaml
---
instance_groups:
- name: postgres
  jobs:
  - name: postgres-server
    release: postgres-release
  - name: database-backuper                       <<
    release: database-backup-and-restore-release  <<
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
  "output_file": "db backup output destination"
}
```

In your release backup script, call database-backuper/bin/backup:

```bash
/var/vcap/jobs/database-backuper/bin/backup /path/to/config.json
```

If using BOSH Backup and Restore, you'll have to copy the database backup to $ARTIFACT_DIRECTORY in your backup script.

```bash
cp <output_file> $ARTIFACT_DIRECTORY 
```

### Supported database adapters
* `postgres`