# Backup and Restore SDK BOSH release

The Backup and Restore SDK BOSH release is used for two distinct things:

1. enabling release authors to incorporate database backup & restore functionality in their releases (See _[Database Backup and Restore](docs/database-backup-restore.md)_)
1. enabling operators to configure their deployments which use external blobstores to be backed up and restored by [BBR](https://github.com/cloudfoundry-incubator/bosh-backup-and-restore) (See _[Blobstore Backup and Restore](docs/blobstore-backup-restore.md)_)

**Slack**: #bbr on https://slack.cloudfoundry.org

**Pivotal Tracker**: https://www.pivotaltracker.com/n/projects/1662777

## CI Status

Backup and Restore SDK Release status [![Build SDK Release Badge](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/create-release/badge)](https://ci.cryo.cf-app.com/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release)

## Database Backup and Restore

| Name     | Versions                 | Status                                                                                                                                                                                                                                                                             |
|:---------|:-------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| MariaDB  | 10.2.x            | [![MariaDB Badge](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/mariadb-system-tests-rds/badge)](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/mariadb-system-tests-rds)  [![MariaDB Badge](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/mariadb-system-tests/badge)](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/mariadb-system-tests)        |
| MySQL    | 5.7.x      | [![MySQL Badge](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/mysql-system-tests-gcp/badge)](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/mysql-system-tests-gcp) (GCP)  [![MySQL Badge](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/mysql-system-tests-rds/badge)](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/mysql-system-tests-rds) (AWS RDS)         |
| Postgres | 9.4.x, 9.6.x, 10.x, 11.x | [![Postgres Badge](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/postgres-system-tests/badge)](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/postgres-system-tests) |

## Blobstore Backup and Restore

### Supported Blobstores

| Name                 | Status                                                                                                                                                                                                                                                                                                          |
|:---------------------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| S3-Compatible        | [![S3 Badge](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/s3-blobstore-backuper-system-tests/badge)](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/s3-blobstore-backuper-system-tests)          |
| Azure                | [![Azure Badge](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/azure-blobstore-backuper-system-tests/badge)](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/azure-blobstore-backuper-system-tests) |
| Google Cloud Storage | [![GCS Badge](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/s3-blobstore-backuper-system-tests/badge)](https://ci.cryo.cf-app.com/api/v1/teams/bosh-backup-restore/pipelines/backup-and-restore-sdk-release/jobs/gcs-blobstore-backuper-system-tests)        |

## Developing

This repository using develop as the main branch, tested releases are tagged with their versions, and master branch represents the latest release.
