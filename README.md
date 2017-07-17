# Backup and Restore SDK BOSH release

A SDK BOSH release used to backup and restore databases for BOSH deployed Cloud Foundry components.

## Database Backup and Restore

### The backup-and-restore instance group

You should co-locate the `database-backup-restorer` job and your release backup scripts on to a backup-and-restore instance in your Cloud Foundry deployment:

[Release Author Guide](http://www.boshbackuprestore.io/bosh-backup-and-restore/release_author_guide.html).

Example BOSH v2 deployment manifest:
```yaml
...
instance_groups:
- name: backup
  networks:
  - name: my-network
  persistent_disk_type: 10GB
  stemcell: default
  update:
    serial: true
  vm_type: m3.large
  azs: [z1]
  instances: 1
  jobs:
  - name: backup-my-release
    properties:
      mydb:
      address: mydb.example.com
      db_scheme: mysql
      port: 3306
    release: my_release
  - name: backup-and-restore-utility
    release: backup-and-restore-sdk-release
...
```

Template a `config.json` as follows:

```json
{
  "username": "db user",
  "password": "db password",
  "host": "db host",
  "port": your_db_port,
  "adapter": "db adapter; see 'Supported database adapters'",
  "database": "name of database to back up"
}
```

#### Supported Database Adapters
* `postgres`
* `mysql`

### Usage

In your release backup script, call `database-backup-restorer/bin/backup`:

```bash
/var/vcap/jobs/database-backup-restorer/bin/backup --config /path/to/config.json --artifact-file $BBR_ARTIFACT_DIRECTORY/artifactFile
```

In your release restore script, call `database-backup-restorer/bin/restore`:

```bash
/var/vcap/jobs/database-backup-restorer/bin/restore --config /path/to/config.json --artifact-file $BBR_ARTIFACT_DIRECTORY/artifactFile
```
