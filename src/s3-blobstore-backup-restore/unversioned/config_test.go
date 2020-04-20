package unversioned_test

import (
	"errors"
	"fmt"

	"s3-blobstore-backup-restore/unversioned"

	"s3-blobstore-backup-restore/incremental/fakes"
	unversionedFakes "s3-blobstore-backup-restore/unversioned/fakes"

	"s3-blobstore-backup-restore/s3bucket"

	"s3-blobstore-backup-restore/incremental"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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

		newBucket = func(bucketName, bucketRegion, endpoint string, accessKey s3bucket.AccessKey, useIAMProfile, _ bool) (unversioned.Bucket, error) {
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

		DescribeTable("passes the appropriate path/vhost information from the config to the bucket builder", func(forcePathStyle bool) {
			config := map[string]unversioned.UnversionedBucketConfig{
				"bucket": unversioned.UnversionedBucketConfig{ForcePathStyle: forcePathStyle},
			}

			forcePathStyles := []bool{}
			newBucketSpy := func(_, _, _ string, _ s3bucket.AccessKey, _, forcePathStyle bool) (unversioned.Bucket, error) {
				forcePathStyles = append(forcePathStyles, forcePathStyle)
				return fakeLiveBucket1, nil
			}

			_, err := unversioned.BuildBackupsToStart(config, newBucketSpy)
			Expect(err).NotTo(HaveOccurred())
			Expect(forcePathStyles[0]).To(Equal(forcePathStyle), "forcePathStyle param to newBucket for live bucket should match bucket config")
			Expect(forcePathStyles[1]).To(Equal(forcePathStyle), "forcePathStyle param to newBucket for backup bucket should match bucket config")
		},
			Entry("we force the path style", true),
			Entry("we allow vhost style", false),
		)

		Context("when bucket initialisation fails", func() {
			var (
				newBucketFails unversioned.NewBucket
				bucketToFail   string
			)

			JustBeforeEach(func() {
				newBucketFails = func(bucketName, bucketRegion, endpoint string, accessKey s3bucket.AccessKey, useIAMProfile, forcePathStyle bool) (unversioned.Bucket, error) {
					if bucketName == bucketToFail {
						return nil, errors.New("oups")
					} else {
						return newBucket(bucketName, bucketRegion, endpoint, accessKey, useIAMProfile, forcePathStyle)
					}
				}
			})

			Context("and it's a live bucket", func() {
				BeforeEach(func() {
					bucketToFail = "live-name2"
				})
				It("returns an error", func() {
					_, err := unversioned.BuildBackupsToStart(configs, newBucketFails)
					Expect(err).To(MatchError("oups"))
				})
			})

			Context("and it's a backup bucket", func() {
				BeforeEach(func() {
					bucketToFail = "backup-name1"
				})
				It("returns an error", func() {
					_, err := unversioned.BuildBackupsToStart(configs, newBucketFails)
					Expect(err).To(MatchError("oups"))
				})
			})
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

		Context("when it fails to check if a live bucket is versioned", func() {
			BeforeEach(func() {
				fakeLiveBucket1.IsVersionedReturns(false, errors.New("oopps"))
			})

			It("errors", func() {
				_, err := unversioned.BuildBackupsToStart(configs, newBucket)
				Expect(err).To(MatchError("oopps"))
			})
		})

		Context("when it fails to check if a backup bucket is versioned", func() {
			BeforeEach(func() {
				fakeBackupBucket1.IsVersionedReturns(false, errors.New("oopps"))
			})

			It("errors", func() {
				_, err := unversioned.BuildBackupsToStart(configs, newBucket)
				Expect(err).To(MatchError("oopps"))
			})
		})

		Context("when a live bucket is versioned", func() {
			BeforeEach(func() {
				fakeLiveBucket1.IsVersionedReturns(true, nil)
			})

			It("errors", func() {
				_, err := unversioned.BuildBackupsToStart(configs, newBucket)
				Expect(err).To(MatchError(fmt.Errorf("bucket %s is versioned", fakeLiveBucket1.Name())))
			})
		})

		Context("when a backup bucket is versioned", func() {
			BeforeEach(func() {
				fakeBackupBucket1.IsVersionedReturns(true, nil)
			})

			It("errors", func() {
				_, err := unversioned.BuildBackupsToStart(configs, newBucket)
				Expect(err).To(MatchError(fmt.Errorf("bucket %s is versioned", fakeBackupBucket1.Name())))
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
				unversioned.NewUnversionedBucket,
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
				unversioned.NewUnversionedBucket,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(backupsToComplete).To(HaveLen(2))
			Expect(backupsToComplete["bucket2"]).To(Equal(incremental.BackupToComplete{
				SameAsBucketID: "bucket1",
			}))
		})

		DescribeTable("passes the appropriate path/vhost information from the config to the bucket builder", func(forcePathStyle bool) {
			config := map[string]unversioned.UnversionedBucketConfig{
				"bucket": unversioned.UnversionedBucketConfig{ForcePathStyle: forcePathStyle},
			}
			existingBlobsArtifact.LoadReturns(map[string]incremental.Backup{
				"bucket": {},
			}, nil)

			forcePathStyles := []bool{}
			newBucketSpy := func(_, _, _ string, _ s3bucket.AccessKey, _, forcePathStyle bool) (unversioned.Bucket, error) {
				forcePathStyles = append(forcePathStyles, forcePathStyle)
				return fakeLiveBucket1, nil
			}

			_, err := unversioned.BuildBackupsToComplete(config, existingBlobsArtifact, newBucketSpy)
			Expect(err).NotTo(HaveOccurred())
			Expect(forcePathStyles).To(HaveLen(1))
			Expect(forcePathStyles[0]).To(Equal(forcePathStyle), "forcePathStyle param to newBucket for live bucket should match bucket config")
		},
			Entry("we force the path style", true),
			Entry("we allow vhost style", false),
		)

		It("returns error when it cannot load existing blobs artifact", func() {
			existingBlobsArtifact.LoadReturns(nil, errors.New("fake load error"))

			_, err := unversioned.BuildBackupsToComplete(
				configs,
				existingBlobsArtifact,
				unversioned.NewUnversionedBucket,
			)

			Expect(err).To(MatchError(ContainSubstring("fake load error")))
		})

		It("returns error when a configured bucketID is not in the existing blobs artifact", func() {
			existingBlobsArtifact.LoadReturns(map[string]incremental.Backup{}, nil)

			_, err := unversioned.BuildBackupsToComplete(
				configs,
				existingBlobsArtifact,
				unversioned.NewUnversionedBucket,
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
			artifact.LoadReturns(map[string]incremental.Backup{
				"bucket1": {
					BucketName:             "backup-name1",
					BucketRegion:           "backup-region1",
					SrcBackupDirectoryPath: "destination-backup-dir1",
				},
				"bucket2": {
					BucketName:             "backup-name2",
					BucketRegion:           "backup-region2",
					SrcBackupDirectoryPath: "destination-backup-dir2",
				},
			}, nil)
		})

		It("builds restore bucket pairs from a config and a backup artifact", func() {
			restoreBucketPairs, err := unversioned.BuildRestoreBucketPairs(configs, artifact, newBucket)

			Expect(err).NotTo(HaveOccurred())

			Expect(restoreBucketPairs).To(Equal(
				map[string]incremental.RestoreBucketPair{
					"bucket1": {
						ConfigLiveBucket:     fakeLiveBucket1,
						ArtifactBackupBucket: fakeBackupBucket1,
					},
					"bucket2": {
						ConfigLiveBucket:     fakeLiveBucket2,
						ArtifactBackupBucket: fakeBackupBucket2,
					},
				},
			))
		})

		DescribeTable("passes the appropriate path/vhost information from the config to the bucket builder", func(forcePathStyle bool) {
			config := map[string]unversioned.UnversionedBucketConfig{
				"bucket": unversioned.UnversionedBucketConfig{ForcePathStyle: forcePathStyle},
			}
			artifact.LoadReturns(map[string]incremental.Backup{
				"bucket": {},
			}, nil)

			forcePathStyles := []bool{}
			newBucketSpy := func(_, _, _ string, _ s3bucket.AccessKey, _, forcePathStyle bool) (unversioned.Bucket, error) {
				forcePathStyles = append(forcePathStyles, forcePathStyle)
				return fakeLiveBucket1, nil
			}

			_, err := unversioned.BuildRestoreBucketPairs(config, artifact, newBucketSpy)
			Expect(err).NotTo(HaveOccurred())
			Expect(forcePathStyles).To(HaveLen(2))
			Expect(forcePathStyles[0]).To(Equal(forcePathStyle), "forcePathStyle param to newBucket for live bucket should match bucket config")
			Expect(forcePathStyles[1]).To(Equal(forcePathStyle), "forcePathStyle param to newBucket for backup bucket should match bucket config")
		},
			Entry("we force the path style", true),
			Entry("we allow vhost style", false),
		)

		Context("when checking if the bucket is versioned", func() {

			Context("when it fails to check if a live bucket is versioned", func() {
				BeforeEach(func() {
					fakeLiveBucket1.IsVersionedReturns(false, errors.New("oopps"))
				})

				It("errors", func() {
					_, err := unversioned.BuildRestoreBucketPairs(configs, artifact, newBucket)
					Expect(err).To(MatchError("oopps"))
				})
			})

			Context("when it fails to check if a backup bucket is versioned", func() {
				BeforeEach(func() {
					fakeBackupBucket1.IsVersionedReturns(false, errors.New("oopps"))
				})

				It("errors", func() {
					_, err := unversioned.BuildRestoreBucketPairs(configs, artifact, newBucket)
					Expect(err).To(MatchError("oopps"))
				})
			})

			Context("when a live bucket is versioned", func() {
				BeforeEach(func() {
					fakeLiveBucket1.IsVersionedReturns(true, nil)
				})

				It("errors", func() {
					_, err := unversioned.BuildRestoreBucketPairs(configs, artifact, newBucket)
					Expect(err).To(MatchError(fmt.Errorf("bucket %s is versioned", fakeLiveBucket1.Name())))
				})
			})

			Context("when a backup bucket is versioned", func() {
				BeforeEach(func() {
					fakeBackupBucket1.IsVersionedReturns(true, nil)
				})

				It("errors", func() {
					_, err := unversioned.BuildRestoreBucketPairs(configs, artifact, newBucket)
					Expect(err).To(MatchError(fmt.Errorf("bucket %s is versioned", fakeBackupBucket1.Name())))
				})
			})
		})

		Context("when bucket initialisation fails", func() {
			var (
				newBucketFails unversioned.NewBucket
				bucketToFail   string
			)

			JustBeforeEach(func() {
				newBucketFails = func(bucketName, bucketRegion, endpoint string, accessKey s3bucket.AccessKey, useIAMProfile, forcePathStyle bool) (unversioned.Bucket, error) {
					if bucketName == bucketToFail {
						return nil, errors.New("oups")
					} else {
						return newBucket(bucketName, bucketRegion, endpoint, accessKey, useIAMProfile, forcePathStyle)
					}
				}
			})

			Context("and it's a live bucket", func() {
				BeforeEach(func() {
					bucketToFail = "live-name2"
				})
				It("returns an error", func() {
					_, err := unversioned.BuildRestoreBucketPairs(configs, artifact, newBucketFails)
					Expect(err).To(MatchError("oups"))
				})
			})

			Context("and it's a backup bucket", func() {
				BeforeEach(func() {
					bucketToFail = "backup-name1"
				})
				It("returns an error", func() {
					_, err := unversioned.BuildRestoreBucketPairs(configs, artifact, newBucketFails)
					Expect(err).To(MatchError("oups"))
				})
			})
		})

		Context("when there are duplicate live buckets", func() {
			BeforeEach(func() {
				fakeLiveBucket2.NameReturns("live-name1")
			})

			It("builds restore bucket pairs marked same", func() {
				artifact.LoadReturns(map[string]incremental.Backup{
					"bucket1": {
						BucketName:             "backup-name1",
						BucketRegion:           "backup-region1",
						SrcBackupDirectoryPath: "destination-backup-dir1",
					},
					"bucket2": {
						SameBucketAs: "bucket1",
					},
				}, nil)

				restoreBucketPairs, err := unversioned.BuildRestoreBucketPairs(configs, artifact, newBucket)
				Expect(err).NotTo(HaveOccurred())

				Expect(restoreBucketPairs).To(Equal(
					map[string]incremental.RestoreBucketPair{
						"bucket1": {
							ConfigLiveBucket:     fakeLiveBucket1,
							ArtifactBackupBucket: fakeBackupBucket1,
						},
						"bucket2": {
							SameAsBucketID: "bucket1",
						},
					},
				))
			})
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
