# Backup and Restore SDK BOSH release

A SDK BOSH release used to backup and restore databases for BOSH deployed Cloud Foundry components.

* [Release Author Guide](https://docs.pivotal.io/tiledev/bbr-devguide.html)

## Why?

Release authors wanting to write backup and restore scripts frequently need to back up and restore databases (or parts of databases).

Rather than have every team figure out the vagaries of backing up all the different kinds of database supported by CF, we've done it for you. The **Backup and Restore SDK** abstracts away the differences between databases, offering a consistent interface for your backup and restore scripts to use.

Behind the scenes, the SDK parses a configuration file passed to it, which selects the appropriate database backup/restore strategy (e.g. `pg_dump` or `mysql` at the required version) and places the backup artifact in the specified location.

## Deploying

### Deploying with `cf-deployment`

Users of [cf-deployment](https://github.com/cloudfoundry/cf-deployment) can simply apply the [backup-restore opsfile](https://github.com/cloudfoundry/cf-deployment/blob/master/operations/experimental/enable-backup-restore.yml). This will deploy the `database-backup-restorer` job on a backup restore VM alongside Cloud Foundry.

### Deploying as an instance group

You should co-locate the `database-backup-restorer` job and your release backup scripts on the same VM. If you use a dedicated backup-and-restore VM instance, co-locate them together on that VM. BOSH Lite is supported for testing.

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

## Usage

### 1. Template `config.json`

Your job should template a `config.json` as follows:

```json
{
  "username": "db user",
  "password": "db password",
  "host": "db host",
  "port": 3306,
  "adapter": "db adapter; see 'Supported database adapters'",
  "database": "name of database to back up",
}
```

**Note**: all fields in `config.json` need to be strings except `port` which is an int.

`tables` is an optional field (a list of strings). If you don't specify it, the entire database will be backed up and restored. If you specify a list of tables, only the tables in that list will be included in the backup, and on restore the other tables in the database will be left as is. If the field is specified and empty, the utility will fail. If the field contains non-existent tables the utility will fail.

```json
{
  "username": "db user",
  "password": "db password",
  "host": "db host",
  "port": 3306,
  "adapter": "db adapter; see 'Supported database adapters'",
  "database": "name of database to back up",
  "tables": ["list", "of", "tables", "to", "back", "up"]
}
```

We have not tested this with foreign key relationships or triggers spanning between tables specified in the `tables` list and other tables in the database not listed there. It's possible those relationships would be lost on restore.

An example of templating using BOSH Links can be seen in the [cf networking release](https://github.com/cloudfoundry-incubator/cf-networking-release/blob/647f7a71b442c25ec29b1cc6484410946f41935c/jobs/bbr-cfnetworkingdb/templates/config.json.erb).

#### Supported Database Adapters

* `postgres` (auto-detects versions between 9.4.x and 9.6.x)
* `mysql`

Note that these have been tested with internal databases only.

### 2. Write scripts to call the SDK binaries

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

### Developing
This repository using master as the main branch, tested releases are tagged with their versions.
