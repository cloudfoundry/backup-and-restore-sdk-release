# Backup and Restore SDK BOSH release

The Backup and Restore SDK BOSH release is used for two distinct things:

1. enabling release authors to incorporate database backup & restore functionality in their releases
1. enabling operators to configure their deployments which use external blobstores to be backed up and restored by [BBR](https://github.com/cloudfoundry-incubator/bosh-backup-and-restore)

**Docs**: [Release Author Guide](http://docs.cloudfoundry.org/bbr/bbr-devguide.html)

**Slack**: #bbr on https://slack.cloudfoundry.org

**Pivotal Tracker**: https://www.pivotaltracker.com/n/projects/1662777

## CI Status

Backup and Restore SDK Release status [![Build SDK Release Badge](https://backup-and-restore.ci.cf-app.com/api/v1/teams/main/pipelines/backup-and-restore-sdk-release/jobs/create-release/badge)](https://backup-and-restore.ci.cf-app.com/teams/main/pipelines/backup-and-restore-sdk-release)

## Developing

This repository using master as the main branch, tested releases are tagged with their versions.

## Incorporating database backups in your release

### Supported Databases

| Name     | Version | Status                                                                                                                                                                                                                                                                                     |
|:---------|:--------|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| MariaDB  | 10.1.x  | [![MariaDB Badge](https://backup-and-restore.ci.cf-app.com/api/v1/teams/main/pipelines/backup-and-restore-sdk-release/jobs/mariadb-system-tests/badge)](https://backup-and-restore.ci.cf-app.com/teams/main/pipelines/backup-and-restore-sdk-release/jobs/mariadb-system-tests)            |
| MySQL    | 5.5.x   | [![MySQL Badge](https://backup-and-restore.ci.cf-app.com/api/v1/teams/main/pipelines/backup-and-restore-sdk-release/jobs/rds-mysql-5.5-system-tests/badge)](https://backup-and-restore.ci.cf-app.com/teams/main/pipelines/backup-and-restore-sdk-release/jobs/rds-mysql-5.5-system-tests)  |
| MySQL    | 5.6.x   | [![MySQL Badge](https://backup-and-restore.ci.cf-app.com/api/v1/teams/main/pipelines/backup-and-restore-sdk-release/jobs/rds-mysql-5.6-system-tests/badge)](https://backup-and-restore.ci.cf-app.com/teams/main/pipelines/backup-and-restore-sdk-release/jobs/rds-mysql-5.6-system-tests)  |
| MySQL    | 5.7.x   | [![MySQL Badge](https://backup-and-restore.ci.cf-app.com/api/v1/teams/main/pipelines/backup-and-restore-sdk-release/jobs/rds-mysql-5.7-system-tests/badge)](https://backup-and-restore.ci.cf-app.com/teams/main/pipelines/backup-and-restore-sdk-release/jobs/rds-mysql-5.7-system-tests)  |
| Postgres | 9.4.x   | [![Postgres Badge](https://backup-and-restore.ci.cf-app.com/api/v1/teams/main/pipelines/backup-and-restore-sdk-release/jobs/postgres-system-tests-9.4/badge)](https://backup-and-restore.ci.cf-app.com/teams/main/pipelines/backup-and-restore-sdk-release/jobs/postgres-system-tests-9.4) |
| Postgres | 9.6.x   | [![Postgres Badge](https://backup-and-restore.ci.cf-app.com/api/v1/teams/main/pipelines/backup-and-restore-sdk-release/jobs/postgres-system-tests-9.6/badge)](https://backup-and-restore.ci.cf-app.com/teams/main/pipelines/backup-and-restore-sdk-release/jobs/postgres-system-tests-9.6) |

### Why?

Release authors wanting to write backup and restore scripts frequently need to back up and restore databases (or parts of databases).

Rather than have every team figure out the vagaries of backing up all the different kinds of database supported by CF, we've done it for you. The **Backup and Restore SDK** abstracts away the differences between databases, offering a consistent interface for your backup and restore scripts to use.

Behind the scenes, the SDK parses a configuration file passed to it, which selects the appropriate database backup/restore strategy (e.g. `pg_dump` or `mysql` at the required version) and places the backup artifact in the specified location.

### Config options

The SDK accepts a JSON document with the following fields:

| name                  | type         | Optional | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
|:----------------------|:-------------|:---------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| username              | string       | no       | Database connection username                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| password              | string       | no       | Database connection password                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| host                  | string       | no       | Database connection host                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| port                  | integer      | no       | Database connection port, no defaulting is done, always needs to be specified                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| adapter               | string       | no       | Database adapter, see [Supported database adapters](#supported-database-adapters)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| database              | string       | no       | Name of the database to backup/restore                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| tables                | string array | yes      | If not specified, the entire database will be backed up/restored. If specified only the tables in that list will be included in the backup, and on restore the other tables in the database will be left as is. If the field is specified and empty, the utility will fail. If the field contains non-existent tables the utility will fail. We have not tested this with foreign key relationships or triggers spanning between tables specified in the `tables` list and other tables in the database not listed there. It's possible those relationships would be lost on restore. |
| tls.skip_host_verify  | bool         | yes      | Skip host verification for Server CA certificate. This needs to be set to `true` if your database is hosted on GCP, as GCP does not support hostname verification.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| tls.cert.ca          | string       | yes      | Server CA certificate. This must be included if any of the `tls` block is specified                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| tls.cert.certificate | string       | yes      | Client certificate for Mutual TLS. This must be specified if `tls.cert.private_key` is given. You will not be able to use this option if your database is hosted on RDS as RDS does not support mutual TLS.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| tls.cert.private_key | string       | yes      | Client private key for Mutual TLS, this must be specified if `tls.cert.certificate` is given.  You will not be able to use this option if your database is hosted on RDS as RDS does not support mutual TLS.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |

#### Supported Database Adapters

* `postgres` (auto-detects versions between `9.4.x` and `9.6.x`)
* `mysql` (auto-detects `MariaDB 10.1.x`, and `MySQL 5.5.x`, `5.6.x`, `5.7.x`. Any other `mysql` variants are not tested)

### Deploying

#### Deploying with `cf-deployment`

Users of [cf-deployment](https://github.com/cloudfoundry/cf-deployment) can simply apply the [backup-restore opsfiles](https://github.com/cloudfoundry/cf-deployment/blob/master/operations/backup-and-restore). This will deploy the `database-backup-restorer` job on a backup restore VM alongside Cloud Foundry.

#### Deploying as an instance group

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

### Usage from another BOSH job

#### 1. Template `config.json`

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

Or if you want to operate on specific tables:

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

For the full list of `config.json` properties see [Config options](#config-options).

An example of templating using BOSH Links can be seen in the [cf networking release](https://github.com/cloudfoundry-incubator/cf-networking-release/blob/647f7a71b442c25ec29b1cc6484410946f41935c/jobs/bbr-cfnetworkingdb/templates/config.json.erb).

#### 2. Write scripts to call the SDK binaries

In your release backup script, call `database-backup-restorer/bin/backup`:

```bash
/var/vcap/jobs/database-backup-restorer/bin/backup --config /path/to/config.json --artifact-file $BBR_ARTIFACT_DIRECTORY/artifactFile
```

In your release restore script, call `database-backup-restorer/bin/restore`:

```bash
/var/vcap/jobs/database-backup-restorer/bin/restore --config /path/to/config.json --artifact-file $BBR_ARTIFACT_DIRECTORY/artifactFile
```

The `restore` script will assume that the database schema has already been created, and matches the one of the backup. For BOSH releases, this usually means `restore` can be called after a successful deploy of the release, at the same version as the backup was taken.

### Usage with [bbr](https://github.com/cloudfoundry-incubator/bosh-backup-and-restore)

For an example of the SDK being used in a release that can be backed up by BBR see the [exemplar release](https://github.com/cloudfoundry-incubator/exemplar-backup-and-restore-release).

## Incorporating external blobstore backups in your deployment

BBR supports the backup and restore of blobstores stored in Amazon S3 buckets and in S3-compatible buckets.

When restoring to a bucket, BBR will only modify blobs that are recorded in the backup artifact. Any other blobs will not be affected.

### Supported Blobstores

| Name         | Status                                                                                                                                                                                                                                                                                                 |
|:-------------|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| S3-Compatible | [![S3 Badge](https://backup-and-restore.ci.cf-app.com/api/v1/teams/main/pipelines/backup-and-restore-sdk-release/jobs/s3-blobstore-backuper-system-tests/badge)](https://backup-and-restore.ci.cf-app.com/teams/main/pipelines/backup-and-restore-sdk-release/jobs/s3-blobstore-backuper-system-tests) |

### S3-Compatible Unversioned Blobstores

External blobstores are backed up by copying blobs to a backup bucket. The unversioned backup-restorer uses the blobstore's copy functionality to transfer blobs between the buckets to avoid transfering and storing the blobs on your instances.

**Restore only works if the backup bucket still exists**. For this reason, you can configure backup buckets to be in a different region.

### S3-Compatible Versioned Blobstores
The versioned backup-restorer only supports S3-compatible buckets that are versioned and support AWS Signature Version 4. For more details about enabling versioning on your blobstore, see the [Cloud Foundry documentation](https://docs.cloudfoundry.org/bbr/external-blobstores.html#enable-versioning).

External blobstores are backed up by storing the current version of each blob, not the actual files. Those versions will be set to be the current versions at restore time. This makes backups and restores faster, but also means that **restores only work if the original bucket still exists**. For more information, see the [Cloud Foundry documentation](https://docs.cloudfoundry.org/bbr/external-blobstores.html).

### Deploying

#### Deploying with `cf-deployment`

`cf-deployment` includes ops files for enabling the backup and restore of external blobstores. See the [Cloud Foundry documentation](https://docs.cloudfoundry.org/bbr/external-blobstores.html#enable-backup-and-restore) for more details.

#### Adding the SDK to your deployment manifest

The S3-compatible blobstore backup and restore scripts are contained in the `s3-versioned-blobstore-backup-restorer` and `s3-unversioned-blobstore-backup-restorer` jobs. Locate one of the jobs on any of the instance groups in your deployment that has a persistent disk (i.e. the `/var/vcap/store` folder should exist).

##### S3 Versioned Properties

The `s3-versioned-blobstore-backup-restorer` job can be configured using the following properties:

* `enabled` [Boolean]: enables the backup and restore scripts. `true` by default.
* `buckets` [Hash]: a map from bucket identifiers to their configuration. For each bucket, you'll need to specify the following properties:
  * `name` [String]: the bucket name
  * `region` [String]: the bucket region
  * `aws_access_key_id` [String]: the AWS access key ID for the bucket
  * `aws_secret_access_key` [String]: the AWS secret access key for the bucket
  * `endpoint` [String]: the endpoint for your storage server, only needed if you are not using AWS S3

Here are example job properties to configure two S3 buckets: `my_bucket` and `other_bucket`.

```yaml
properties:
  enabled: true
  buckets:
    my_bucket:
      name: "((my_bucket_key))"
      region: "((aws_region))"
      aws_access_key_id: "((my_bucket_access_key_id))"
      aws_secret_access_key: "((my_bucket_secret_access_key))"
    other_bucket:
      name: "((other_bucket_package_directory_key))"
      region: "((aws_region))"
      aws_access_key_id: "((other_bucket_access_key_id))"
      aws_secret_access_key: "((other_bucket_secret_access_key))"
```

##### S3 Unversioned Properties

The `s3-unversioned-blobstore-backup-restorer` job can be configured using the following properties:

* `enabled` [Boolean]: enables the backup and restore scripts. `true` by default.
* `buckets` [Hash]: a map from bucket identifiers to their configuration. For each bucket, you'll need to specify the following properties:
  * `name` [String]: the bucket name
  * `region` [String]: the bucket region
  * `aws_access_key_id` [String]: the AWS access key ID for the bucket
  * `aws_secret_access_key` [String]: the AWS secret access key for the bucket
  * `endpoint` [String]: the endpoint for your storage server, only needed if you are not using AWS S3
  * `backup` [Object]: the backup bucket configuration
    * `name` [String]: the backup bucket name
    * `region` [String]: the backup bucket region

Here are example job properties to configure two S3 buckets: `my_bucket` and `other_bucket`.

```yaml
properties:
  enabled: true
  buckets:
    my_bucket:
      name: "((my_bucket_key))"
      region: "((aws_region))"
      aws_access_key_id: "((my_bucket_access_key_id))"
      aws_secret_access_key: "((my_bucket_secret_access_key))"
      backup:
        name: "((my_bucket_backup_key))"
        region: "((aws_backup_region))"
    other_bucket:
      name: "((other_bucket_package_directory_key))"
      region: "((aws_region))"
      aws_access_key_id: "((other_bucket_access_key_id))"
      aws_secret_access_key: "((other_bucket_secret_access_key))"
      backup:
        name: "((other_bucket_backup_key))"
        region: "((aws_backup_region))"
```
