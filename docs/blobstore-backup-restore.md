## Blobstore Backup and Restore

### Usage

Three external blobstore types are supported
- [S3-Compatible Unversioned Blobstores](#S3-Compatible-Unversioned-Blobstores)
- [S3-Compatible Versioned Blobstores](#S3-Compatible-Versioned-Blobstores)
- [Azure Blobstores](#Azure-Blobstores)

Locate one of the jobs on any of the instance groups in your deployment that has a persistent disk (i.e. the `/var/vcap/store` folder should exist). When deploying with `cf-deployment`, refer to ops files for enabling the backup and restore of external blobstores. See the [Cloud Foundry documentation](https://docs.cloudfoundry.org/bbr/external-blobstores.html#enable-backup-and-restore) for more details.

### S3-Compatible Unversioned Blobstores

External blobstores are backed up by copying blobs to a backup bucket. The unversioned backup-restorer uses the blobstore's copy functionality to transfer blobs between the buckets to avoid transfering and storing the blobs on your instances.

**Restore only works if the backup bucket still exists**. For this reason, you can configure backup buckets to be in a different region.

##### S3 Unversioned Properties

The `s3-unversioned-blobstore-backup-restorer` job can be configured using the following properties:

* `enabled` [Boolean]: enables the backup and restore scripts. `true` by default.
* `buckets` [Hash]: a map from bucket identifiers to their configuration. For each bucket, you'll need to specify the following properties:
  * `name` [String]: the bucket name
  * `region` [String]: the bucket region
  * `aws_access_key_id` [String]: the AWS access key ID for the bucket
  * `aws_secret_access_key` [String]: the AWS secret access key for the bucket
  * `endpoint` [String]: the endpoint for your storage server, only needed if you are not using AWS S3
  * `use_iam_profile` [Boolean]: enable using AWS IAM instance profile to connect to the AWS s3 bucket; default to false
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

### S3-Compatible Versioned Blobstores
The versioned backup-restorer only supports S3-compatible buckets that are versioned and support AWS Signature Version 4. For more details about enabling versioning on your blobstore, see the [Cloud Foundry documentation](https://docs.cloudfoundry.org/bbr/external-blobstores.html#enable-versioning).

External blobstores are backed up by storing the current version of each blob, not the actual files. Those versions will be set to be the current versions at restore time. This makes backups and restores faster, but also means that **restores only work if the original bucket still exists**. For more information, see the [Cloud Foundry documentation](https://docs.cloudfoundry.org/bbr/external-blobstores.html).

##### S3 Versioned Properties

The `s3-versioned-blobstore-backup-restorer` job can be configured using the following properties:

* `enabled` [Boolean]: enables the backup and restore scripts. `true` by default.
* `buckets` [Hash]: a map from bucket identifiers to their configuration. For each bucket, you'll need to specify the following properties:
  * `name` [String]: the bucket name
  * `region` [String]: the bucket region
  * `aws_access_key_id` [String]: the AWS access key ID for the bucket
  * `aws_secret_access_key` [String]: the AWS secret access key for the bucket
  * `endpoint` [String]: the endpoint for your storage server, only needed if you are not using AWS S3
  * `use_iam_profile` [Boolean]: enable using AWS IAM instance profile to connect to the AWS s3 bucket instead of AWS access keys; default to false

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
      use_iam_profile: true
```

### Azure Blobstores
The Azure backup-restorer only supports Azure storage containers that have soft delete enabled. To enable soft

The Azure storage containers are backuped up by storing the `Etags` of each filename, not the actual files. This makes backups and restores faster, but also means that **restores only work if the original container still exists**. For more information, see the [Cloud Foundry documentation](https://docs.cloudfoundry.org/bbr/external-blobstores.html).

