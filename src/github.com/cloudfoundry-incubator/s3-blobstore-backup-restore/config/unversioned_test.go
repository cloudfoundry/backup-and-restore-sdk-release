package config_test

import (
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unversioned", func() {
	Context("BuildIncrementalBackupsToStart", func() {
		It("builds a backup to start from a config", func() {
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

			backupsToStart, err := config.BuildIncrementalBackupsToStart(configs)

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
})
