## Blobstore Backup and Restore

Operators can deploy these jobs with their deployments to back up and restore external blobstores by [BBR](https://github.com/cloudfoundry-incubator/bosh-backup-and-restore).

### Usage

The following external blobstore types are supported:
- [S3-Compatible Unversioned](#S3-Compatible-Unversioned-Blobstores)
- [S3-Compatible Versioned](#S3-Compatible-Versioned-Blobstores)
- [Azure](#Azure-Blobstores)
- [Google Cloud Storage](#Google-Cloud-Storage-Blobstores)

Locate one of the jobs on any instance group in your deployment that has a persistent disk (i.e. the `/var/vcap/store` folder should exist). When deploying with `cf-deployment`, refer to ops files for enabling the backup and restore of external blobstores. See the [Cloud Foundry documentation](https://docs.cloudfoundry.org/bbr/external-blobstores.html#enable-backup-and-restore) for more details.

### S3-Compatible Unversioned Blobstores

Unversioned S3-compatible blobstores are backed up by copying blobs to backup buckets. `s3-unversioned-blobstore-backup-restorer` uses the blobstore's copy functionality to transfer blobs between the buckets to avoid transferring and storing the blobs on your instances. This job only works for S3-compatible blobstores that support AWS Signature Version 4.

**Restore only works if the backup buckets still exist**. For this reason, you can configure backup buckets to be in a different region from your source buckets.

##### S3 Unversioned Properties

The `s3-unversioned-blobstore-backup-restorer` job can be configured using the following properties:

* `enabled` [Boolean]: enables the backup and restore scripts. `false` by default.
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

`s3-versioned-blobstore-backup-restorer` only supports S3-compatible buckets that are versioned and support AWS Signature Version 4. For more details about enabling versioning and retention policy on your blobstore, see the [Cloud Foundry documentation](https://docs.cloudfoundry.org/bbr/external-blobstores.html#enable-s3-versioning).

`s3-versioned-blobstore-backup-restorer` backs up blobstores by storing the current version of each blob, not the actual files. Those versions will be set to be the current versions at restore time. This makes backups and restores faster, but also means that **restores only work if the original buckets still exist**. For more information, see the [Cloud Foundry documentation](https://docs.cloudfoundry.org/bbr/external-blobstores.html).

##### S3 Versioned Properties

The `s3-versioned-blobstore-backup-restorer` job can be configured using the following properties:

* `enabled` [Boolean]: enables the backup and restore scripts. `false` by default.
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

The Azure backup-restorer only supports Azure storage containers that have soft delete enabled.

The Azure storage containers are backed up by storing the `Etag`s of each blob, not the actual blobs. This makes backups and restores faster, but also means that **restores only work if the original containers still exist**.

#### Azure storage properties

The `azure-blobstore-backup-restorer` job can be configured using the following properties:

* `enabled` [Boolean]: enables the backup and restore scripts. `false` by default.
* `containers` [Hash]: a map from container identifiers to their configuration. For each container, you'll need to specify the following properties:
  * `name` [String]: the container name
  * `azure_storage_account` [String]: the Azure storage account name for the container
  * `azure_storage_key` [String]: the Azure storage account key for the container
  * `environment` [String]: the [sovereign cloud](https://www.microsoft.com/en-us/trustcenter/cloudservices/nationalcloud) environment for the container. Accepted values are:
    - `AzureCloud` (the global Azure cloud, default)
    - `AzureChinaCloud`
    - `AzureUSGovernment`
    - `AzureGermanCloud`
  * `restore_from` [Hash]: Optional, only configure when restoring to a destination container that belongs a different azure storage account from the backed up source container.
    - `azure_storage_account` [String]: the azure storage account name for the source container
    - `azure_storage_key` [String]: the azure storage account key for the source container

Here are example job properties to configure two Azure storage containers: `destination_container1` and `destination_container2`.

```yaml
properties:
  enabled: true
  containers:
    destination_container1:
      name: "((destination_container1_name))"
      azure_storage_account: "((destination_container1_storage_account_name))"
      azure_storage_key: "((destination_container1_storage_access_key))"
      restore_from:
        azure_storage_account: "((source_container1_storage_account_name))"
        azure_storage_key: "((source_container1_storage_access_key))"
    destination_container2:
      name: "((destination_container2_name))"
      azure_storage_account: "((destination_container2_storage_account_name))"
      azure_storage_key: "((destination_container2_storage_access_key))"
```

### Google Cloud Storage Blobstores

`gcs-blobstore-backup-restorer` only supports Google Cloud Storage (GCS) buckets that have versioning enabled. For more details about enabling versioning and retention policy on your blobstore, see the [Cloud Foundry documentation](https://docs.cloudfoundry.org/bbr/external-blobstores.html#gcs).

`gcs-blobstore-backup-restorer` backs up blobstores by storing the generation number of each live blob, not the actual files. At restore time, those versions will be restored. This makes backups and restores faster, but also means that **restores only work if the original buckets still exist**. For more information, see the [Cloud Foundry documentation](https://docs.cloudfoundry.org/bbr/external-blobstores.html#gcs).

##### Google Cloud Storage Properties

The `gcs-blobstore-backup-restorer` job can be configured using the following properties:

* `enabled` [Boolean]: enables the backup and restore scripts. `false` by default.
* `buckets` [Hash]: a map from bucket identifiers to their configuration. For each bucket, you'll need to specify the following properties:
  * `name` [String]: the bucket name
  * `gcp_service_account_key` [String]: JSON service account key

Here are example job properties to configure two GCS buckets: `my_bucket` and `other_bucket`.

```yaml
properties:
  enabled: true
  buckets:
    my_bucket:
      name: "((my_bucket_key))"
      gcp_service_account_key: "((my_bucket_gcp_service_account_key))"
    other_bucket:
      name: "((other_bucket_key))"
      gcp_service_account_key: "((other_bucket_gcp_service_account_key))"
```
