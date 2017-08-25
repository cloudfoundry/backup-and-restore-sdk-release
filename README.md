# Backup and Restore SDK BOSH release

A SDK BOSH release used to backup and restore databases for BOSH deployed Cloud Foundry components.

* [Release Author Guide](http://www.boshbackuprestore.io/bosh-backup-and-restore/release_author_guide.html)

## Database Backup and Restore

### The backup-and-restore instance group

You should co-locate the `database-backup-restorer` job and your release backup scripts on the same VM. If you use a dedicated backup-and-restore VM instance, co-locate them together on that VM.

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
  - name: backup-scripts
    properties:
      mydb:
      address: mydb.example.com
      db_scheme: mysql
      port: 3306
    release: my_release
  - name: database-backup-restorer
    release: backup-and-restore-sdk
...
```

Note: if you are using [cf-deployment](https://github.com/cloudfoundry/cf-deployment) then you can use the [backup-restore opsfile](https://github.com/cloudfoundry/cf-deployment/blob/master/operations/experimental/deploy-bosh-backup-restore.yml).

Template a `config.json` as follows:

```json
{
  "username": "db user",
  "password": "db password",
  "host": "db host",
  "port": 3306, 
  "adapter": "db adapter; see 'Supported database adapters'",
  "database": "name of database to back up"
}
```

Note: all fields in `config.json` need to be strings except `port` which is an int.

An example of templating using BOSH Links can be seen in the [cf networking release](https://github.com/cloudfoundry-incubator/cf-networking-release/blob/647f7a71b442c25ec29b1cc6484410946f41935c/jobs/bbr-cfnetworkingdb/templates/config.json.erb).

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

The `restore` script will assume that the database schema has already been created, and matches the one of the backup. For BOSH releases, this usually means `restore` can be called after a successful deploy of the release, at the same version as the backup was taken.

#### Usage with [bbr](https://github.com/cloudfoundry-incubator/bosh-backup-and-restore)

For an example of the sdk being used in a release that can be backed up by bbr see the [exemplar release](https://github.com/cloudfoundry-incubator/exemplar-backup-and-restore-release).
