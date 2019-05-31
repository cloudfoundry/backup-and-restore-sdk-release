package unversioned_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/unversioned"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental/fakes"
	unversionedFakes "github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/unversioned/fakes"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/s3bucket"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unversioned", func() {
	var (
		configs           map[string]unversioned.UnversionedBucketConfig
		bucket1Config     unversioned.UnversionedBucketConfig
		newBucket         unversioned.NewBucket
		fakeLiveBucket1   *unversionedFakes.FakeBucket
		fakeBackupBucket1 *unversionedFakes.FakeBucket
		fakeLiveBucket2   *unversionedFakes.FakeBucket
		fakeBackupBucket2 *unversionedFakes.FakeBucket
	)

	BeforeEach(func() {
		bucket1Config = unversioned.UnversionedBucketConfig{
			Name:               "live-name1",
			Region:             "live-region1",
			AwsAccessKeyId:     "my-id",
			AwsSecretAccessKey: "my-secret-key",
			Endpoint:           "my-s3-endpoint.aws",
			UseIAMProfile:      false,
			Backup: unversioned.BackupBucketConfig{
				Name:   "backup-name1",
				Region: "backup-region1",
			},
		}

		configs = map[string]unversioned.UnversionedBucketConfig{
			"bucket1": bucket1Config,
			"bucket2": {
				Name:               "live-name2",
				Region:             "live-region2",
				AwsAccessKeyId:     "my-id",
				AwsSecretAccessKey: "my-secret-key",
				Endpoint:           "my-s3-endpoint.aws",
				UseIAMProfile:      false,
				Backup: unversioned.BackupBucketConfig{
					Name:   "backup-name2",
					Region: "backup-region2",
				},
			},
		}

		fakeLiveBucket1 = new(unversionedFakes.FakeBucket)
		fakeLiveBucket1.NameReturns("live-name1")

		fakeLiveBucket2 = new(unversionedFakes.FakeBucket)
		fakeLiveBucket2.NameReturns("live-name2")

		fakeBackupBucket1 = new(unversionedFakes.FakeBucket)
		fakeBackupBucket1.NameReturns("backup-name1")

		fakeBackupBucket2 = new(unversionedFakes.FakeBucket)
		fakeBackupBucket2.NameReturns("backup-name2")

		newBucket = func(bucketName, bucketRegion, endpoint string, accessKey s3bucket.AccessKey, useIAMProfile bool) (unversioned.Bucket, error) {
			if endpoint == "my-s3-endpoint.aws" && accessKey.Secret == "my-secret-key" && accessKey.Id == "my-id" && !useIAMProfile {
				if bucketName == "live-name1" && bucketRegion == "live-region1" {
					return fakeLiveBucket1, nil
				} else if bucketName == "live-name2" && bucketRegion == "live-region2" {
					return fakeLiveBucket2, nil
				} else if bucketName == "backup-name1" && bucketRegion == "backup-region1" {
					return fakeBackupBucket1, nil
				} else if bucketName == "backup-name2" && bucketRegion == "backup-region2" {
					return fakeBackupBucket2, nil
				}
			}

			return nil, errors.New("new bucket called with invalid arguments")
		}
	})

	Context("BuildBackupsToStart", func() {
		It("builds backups to start from a config", func() {
			backupsToStart, err := unversioned.BuildBackupsToStart(configs, newBucket)
			Expect(err).NotTo(HaveOccurred())

			Expect(backupsToStart).To(Equal(
				map[string]incremental.BackupToStart{
					"bucket1": {
						BucketPair: incremental.BackupBucketPair{
							ConfigLiveBucket:   fakeLiveBucket1,
							ConfigBackupBucket: fakeBackupBucket1,
						},
						BackupDirectoryFinder: incremental.Finder{},
					},
					"bucket2": {
						BucketPair: incremental.BackupBucketPair{
							ConfigLiveBucket:   fakeLiveBucket2,
							ConfigBackupBucket: fakeBackupBucket2,
						},
						BackupDirectoryFinder: incremental.Finder{},
					},
				},
			))
		})

		Context("when the same bucket is configured for two bucket IDs", func() {
			BeforeEach(func() {
				configs = map[string]unversioned.UnversionedBucketConfig{
					"bucket1": bucket1Config,
					"bucket2": bucket1Config,
				}
				fakeLiveBucket2.NameReturns("live-name1")
			})

			It("builds backups to start", func() {
				backupsToStart, err := unversioned.BuildBackupsToStart(configs, newBucket)

				Expect(err).NotTo(HaveOccurred())

				Expect(backupsToStart).To(Equal(
					map[string]incremental.BackupToStart{
						"bucket1": {
							BucketPair: incremental.BackupBucketPair{
								ConfigLiveBucket:   fakeLiveBucket1,
								ConfigBackupBucket: fakeBackupBucket1,
							},
							BackupDirectoryFinder: incremental.Finder{},
						},
						"bucket2": {
							SameAsBucketID: "bucket1",
						},
					},
				))
			})
		})
	})

	Context("BuildBackupsToComplete", func() {
		var existingBlobsArtifact *fakes.FakeArtifact

		BeforeEach(func() {
			existingBlobsArtifact = new(fakes.FakeArtifact)
		})

		It("builds backups to complete from a config", func() {
			existingBlobsArtifact.LoadReturns(map[string]incremental.Backup{
				"bucket1": {
					SrcBackupDirectoryPath: "source-backup-dir1",
					DstBackupDirectoryPath: "destination-backup-dir1",
					Blobs: []string{
						"source-backup-dir1/blob-path1",
						"source-backup-dir1/blob-path2",
					},
				},
				"bucket2": {
					SrcBackupDirectoryPath: "source-backup-dir2",
					DstBackupDirectoryPath: "destination-backup-dir2",
					Blobs: []string{
						"source-backup-dir2/blob-path1",
						"source-backup-dir2/blob-path2",
					},
				},
			}, nil)

			backupsToComplete, err := unversioned.BuildBackupsToComplete(
				configs,
				existingBlobsArtifact,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(backupsToComplete).To(HaveLen(2))
			Expect(existingBlobsArtifact.LoadCallCount()).To(Equal(1))
			for _, n := range []string{"1", "2"} {
				Expect(backupsToComplete).To(HaveKey("bucket" + n))
				Expect(backupsToComplete["bucket"+n].BackupBucket.Name()).To(Equal("backup-name" + n))
				Expect(backupsToComplete["bucket"+n].BackupBucket.Region()).To(Equal("backup-region" + n))

				Expect(backupsToComplete["bucket"+n].BackupDirectory.Bucket.Name()).To(Equal("backup-name" + n))
				Expect(backupsToComplete["bucket"+n].BackupDirectory.Bucket.Region()).To(Equal("backup-region" + n))

				Expect(backupsToComplete["bucket"+n].BackupDirectory.Path).To(Equal("destination-backup-dir" + n))

				Expect(backupsToComplete["bucket"+n].BlobsToCopy).To(ConsistOf(
					incremental.BackedUpBlob{
						Path:                "source-backup-dir" + n + "/blob-path1",
						BackupDirectoryPath: "source-backup-dir" + n,
					},
					incremental.BackedUpBlob{
						Path:                "source-backup-dir" + n + "/blob-path2",
						BackupDirectoryPath: "source-backup-dir" + n,
					},
				))
			}
		})

		It("builds backups to complete marked same", func() {
			existingBlobsArtifact.LoadReturns(map[string]incremental.Backup{
				"bucket1": {
					SrcBackupDirectoryPath: "source-backup-dir1",
					DstBackupDirectoryPath: "destination-backup-dir1",
					Blobs: []string{
						"source-backup-dir1/blob-path1",
						"source-backup-dir1/blob-path2",
					},
				},
				"bucket2": {
					SameBucketAs: "bucket1",
				},
			}, nil)

			backupsToComplete, err := unversioned.BuildBackupsToComplete(
				configs,
				existingBlobsArtifact,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(backupsToComplete).To(HaveLen(2))
			Expect(backupsToComplete["bucket2"]).To(Equal(incremental.BackupToComplete{
				SameAsBucketID: "bucket1",
			}))
		})

		It("returns error when it cannot load existing blobs artifact", func() {
			existingBlobsArtifact.LoadReturns(nil, errors.New("fake load error"))

			_, err := unversioned.BuildBackupsToComplete(
				configs,
				existingBlobsArtifact,
			)

			Expect(err).To(MatchError(ContainSubstring("fake load error")))
		})

		It("returns error when a configured bucketID is not in the existing blobs artifact", func() {
			existingBlobsArtifact.LoadReturns(map[string]incremental.Backup{}, nil)

			_, err := unversioned.BuildBackupsToComplete(
				configs,
				existingBlobsArtifact,
			)

			Expect(err).To(Or(
				MatchError("failed to find bucket identifier 'bucket1' in buckets config"),
				MatchError("failed to find bucket identifier 'bucket2' in buckets config"),
			))
		})
	})

	Context("BuildRestoreBucketPairs", func() {
		var artifact *fakes.FakeArtifact

		BeforeEach(func() {
			artifact = new(fakes.FakeArtifact)
		})

		It("builds restore bucket pairs from a config and a backup artifact", func() {
			artifact.LoadReturns(map[string]incremental.Backup{
				"bucket1": {
					BucketName:             "backup-artifact-name1",
					BucketRegion:           "backup-artifact-region1",
					SrcBackupDirectoryPath: "destination-backup-dir1",
				},
				"bucket2": {
					BucketName:             "backup-artifact-name2",
					BucketRegion:           "backup-artifact-region2",
					SrcBackupDirectoryPath: "destination-backup-dir2",
				},
			}, nil)

			restoreBucketPairs, err := unversioned.BuildRestoreBucketPairs(configs, artifact, newBucket)

			Expect(err).NotTo(HaveOccurred())
			Expect(restoreBucketPairs).To(HaveLen(2))
			for _, n := range []string{"1", "2"} {
				Expect(restoreBucketPairs).To(HaveKey("bucket" + n))
				Expect(restoreBucketPairs["bucket"+n].ConfigLiveBucket.Name()).To(Equal("live-name" + n))
				Expect(restoreBucketPairs["bucket"+n].ConfigLiveBucket.Region()).To(Equal("live-region" + n))
				Expect(restoreBucketPairs["bucket"+n].ArtifactBackupBucket.Name()).To(Equal("backup-artifact-name" + n))
				Expect(restoreBucketPairs["bucket"+n].ArtifactBackupBucket.Region()).To(Equal("backup-artifact-region" + n))
			}
		})

		It("builds restore bucket pairs marked same", func() {
			artifact.LoadReturns(map[string]incremental.Backup{
				"bucket1": {
					BucketName:             "backup-artifact-name1",
					BucketRegion:           "backup-artifact-region1",
					SrcBackupDirectoryPath: "destination-backup-dir1",
				},
				"bucket2": {
					SameBucketAs: "bucket1",
				},
			}, nil)

			restoreBucketPairs, err := unversioned.BuildRestoreBucketPairs(configs, artifact, newBucket)

			Expect(err).NotTo(HaveOccurred())
			Expect(restoreBucketPairs).To(HaveLen(2))
			Expect(restoreBucketPairs).To(HaveKey("bucket1"))
			Expect(restoreBucketPairs["bucket1"].ConfigLiveBucket.Name()).To(Equal("live-name1"))
			Expect(restoreBucketPairs["bucket1"].ConfigLiveBucket.Region()).To(Equal("live-region1"))
			Expect(restoreBucketPairs["bucket1"].ArtifactBackupBucket.Name()).To(Equal("backup-artifact-name1"))
			Expect(restoreBucketPairs["bucket1"].ArtifactBackupBucket.Region()).To(Equal("backup-artifact-region1"))
			Expect(restoreBucketPairs).To(HaveKey("bucket2"))
			Expect(restoreBucketPairs["bucket2"]).To(Equal(incremental.RestoreBucketPair{
				SameAsBucketID: "bucket1",
			}))
		})

		It("returns error when it cannot load backup artifact", func() {
			artifact.LoadReturns(nil, errors.New("fake load error"))

			_, err := unversioned.BuildRestoreBucketPairs(configs, artifact, newBucket)

			Expect(err).To(MatchError(ContainSubstring("fake load error")))
		})

		It("returns error when the backup artifact does not have a configured bucket ID", func() {
			artifact.LoadReturns(map[string]incremental.Backup{}, nil)

			_, err := unversioned.BuildRestoreBucketPairs(configs, artifact, newBucket)

			Expect(err).To(MatchError(ContainSubstring("backup artifact does not contain bucket ID")))
		})
	})
})
