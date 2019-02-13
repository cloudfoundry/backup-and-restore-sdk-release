package config_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/config"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unversioned", func() {
	configs := map[string]config.UnversionedBucketConfig{
		"bucket1": {
			BucketConfig: config.BucketConfig{
				Name:               "live-name1",
				Region:             "live-region1",
				AwsAccessKeyId:     "my-id",
				AwsSecretAccessKey: "my-secret-key",
				Endpoint:           "my-s3-endpoint.aws",
				UseIAMProfile:      false,
			},
			Backup: config.BackupBucketConfig{
				Name:   "backup-name1",
				Region: "backup-region1",
			},
		},
		"bucket2": {
			BucketConfig: config.BucketConfig{
				Name:               "live-name2",
				Region:             "live-region2",
				AwsAccessKeyId:     "my-id",
				AwsSecretAccessKey: "my-secret-key",
				Endpoint:           "my-s3-endpoint.aws",
				UseIAMProfile:      false,
			},
			Backup: config.BackupBucketConfig{
				Name:   "backup-name2",
				Region: "backup-region2",
			},
		},
	}

	Context("BuildBackupsToStart", func() {
		It("builds backups to start from a config", func() {
			backupsToStart, err := config.BuildBackupsToStart(configs)

			Expect(err).NotTo(HaveOccurred())
			Expect(backupsToStart).To(HaveLen(2))
			for _, n := range []string{"1", "2"} {
				Expect(backupsToStart).To(HaveKey("bucket" + n))
				Expect(backupsToStart["bucket"+n].BucketPair.LiveBucket.Name()).To(Equal("live-name" + n))
				Expect(backupsToStart["bucket"+n].BucketPair.LiveBucket.Region()).To(Equal("live-region" + n))
				Expect(backupsToStart["bucket"+n].BucketPair.BackupBucket.Name()).To(Equal("backup-name" + n))
				Expect(backupsToStart["bucket"+n].BucketPair.BackupBucket.Region()).To(Equal("backup-region" + n))

				Expect(backupsToStart["bucket1"].BackupDirectoryFinder).NotTo(BeNil())
			}
		})
	})

	Context("BuildBackupsToComplete", func() {
		var backupArtifact *fakes.FakeArtifact
		var existingBlobsArtifact *fakes.FakeArtifact

		BeforeEach(func() {
			backupArtifact = new(fakes.FakeArtifact)
			existingBlobsArtifact = new(fakes.FakeArtifact)
		})

		It("builds backups to complete from a config", func() {
			backupArtifact.LoadReturns(map[string]incremental.BucketBackup{
				"bucket1": {
					BucketName:          "backup-name1",
					BucketRegion:        "backup-region1",
					BackupDirectoryPath: "new-backup-dir1",
				},
				"bucket2": {
					BucketName:          "backup-name2",
					BucketRegion:        "backup-region2",
					BackupDirectoryPath: "new-backup-dir2",
				},
			}, nil)
			existingBlobsArtifact.LoadReturns(map[string]incremental.BucketBackup{
				"bucket1": {
					BackupDirectoryPath: "existing-backup-dir1",
					Blobs: []string{
						"blob-path1",
						"blob-path2",
					},
				},
				"bucket2": {
					BackupDirectoryPath: "existing-backup-dir2",
					Blobs: []string{
						"blob-path1",
						"blob-path2",
					},
				},
			}, nil)

			backupsToComplete, err := config.BuildBackupsToComplete(
				configs,
				backupArtifact,
				existingBlobsArtifact,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(backupsToComplete).To(HaveLen(2))
			Expect(backupArtifact.LoadCallCount()).To(Equal(1))
			Expect(existingBlobsArtifact.LoadCallCount()).To(Equal(1))
			for _, n := range []string{"1", "2"} {
				Expect(backupsToComplete).To(HaveKey("bucket" + n))
				Expect(backupsToComplete["bucket"+n].BackupBucket.Name()).To(Equal("backup-name" + n))
				Expect(backupsToComplete["bucket"+n].BackupBucket.Region()).To(Equal("backup-region" + n))

				Expect(backupsToComplete["bucket"+n].BackupDirectory.Bucket.Name()).To(Equal("backup-name" + n))
				Expect(backupsToComplete["bucket"+n].BackupDirectory.Bucket.Region()).To(Equal("backup-region" + n))

				Expect(backupsToComplete["bucket"+n].BackupDirectory.Path).To(Equal("new-backup-dir" + n))

				Expect(backupsToComplete["bucket"+n].BlobsToCopy).To(ConsistOf(
					incremental.BackedUpBlob{
						Path:                "blob-path1",
						BackupDirectoryPath: "existing-backup-dir" + n,
					},
					incremental.BackedUpBlob{
						Path:                "blob-path2",
						BackupDirectoryPath: "existing-backup-dir" + n,
					},
				))
			}
		})

		It("returns error when it cannot load backup artifact", func() {
			backupArtifact.LoadReturns(nil, errors.New("fake load error"))

			_, err := config.BuildBackupsToComplete(
				configs,
				backupArtifact,
				existingBlobsArtifact,
			)

			Expect(err).To(MatchError(ContainSubstring("fake load error")))
		})

		It("returns error when it cannot load existing blobs artifact", func() {
			existingBlobsArtifact.LoadReturns(nil, errors.New("fake load error"))

			_, err := config.BuildBackupsToComplete(
				configs,
				backupArtifact,
				existingBlobsArtifact,
			)

			Expect(err).To(MatchError(ContainSubstring("fake load error")))
		})

		It("returns error when a configured bucketID is not in the existing blobs artifact", func() {
			backupArtifact.LoadReturns(map[string]incremental.BucketBackup{}, nil)
			existingBlobsArtifact.LoadReturns(map[string]incremental.BucketBackup{}, nil)

			_, err := config.BuildBackupsToComplete(
				configs,
				backupArtifact,
				existingBlobsArtifact,
			)

			Expect(err).To(Or(
				MatchError("failed to find bucket identifier 'bucket1' in buckets config"),
				MatchError("failed to find bucket identifier 'bucket2' in buckets config"),
			))
		})
	})

	Context("BuildRestoreBucketPairs", func() {
		var backupArtifact *fakes.FakeArtifact

		BeforeEach(func() {
			backupArtifact = new(fakes.FakeArtifact)
		})

		It("builds restore bucket pairs from a config and a backup artifact", func() {
			backupArtifact.LoadReturns(map[string]incremental.BucketBackup{
				"bucket1": {
					BucketName:          "backup-name1",
					BucketRegion:        "backup-region1",
					BackupDirectoryPath: "new-backup-dir1",
				},
				"bucket2": {
					BucketName:          "backup-name2",
					BucketRegion:        "backup-region2",
					BackupDirectoryPath: "new-backup-dir2",
				},
			}, nil)

			restoreBucketPairs, err := config.BuildRestoreBucketPairs(configs, backupArtifact)

			Expect(err).NotTo(HaveOccurred())
			Expect(restoreBucketPairs).To(HaveLen(2))
			for _, n := range []string{"1", "2"} {
				Expect(restoreBucketPairs).To(HaveKey("bucket" + n))
			}
		})

		It("returns error when it cannot load backup artifact", func() {
			backupArtifact.LoadReturns(nil, errors.New("fake load error"))

			_, err := config.BuildRestoreBucketPairs(configs, backupArtifact)

			Expect(err).To(MatchError(ContainSubstring("fake load error")))
		})
	})
})
