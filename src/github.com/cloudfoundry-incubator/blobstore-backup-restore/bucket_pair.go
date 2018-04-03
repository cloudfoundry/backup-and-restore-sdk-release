package blobstore

//go:generate counterfeiter -o fakes/fake_unversioned_bucket_pair.go . UnversionedBucketPair
type UnversionedBucketPair interface {
	Backup(backupLocation string) (BackupBucketAddress, error)
}

type S3BucketPair struct {
	LiveBucket   UnversionedBucket
	BackupBucket UnversionedBucket
}

func NewS3BucketPair(liveBucketName, liveBucketRegion, endpoint string, accessKey S3AccessKey,
	backupBucketName string, backupBucketRegion string) (S3BucketPair, error) {

	liveS3Client, err := newS3Client(liveBucketRegion, endpoint, accessKey)
	if err != nil {
		return S3BucketPair{}, err
	}

	backupS3Client, err := newS3Client(backupBucketRegion, endpoint, accessKey)
	if err != nil {
		return S3BucketPair{}, err
	}

	return S3BucketPair{
		LiveBucket: S3UnversionedBucket{
			S3Bucket{
				name:       liveBucketName,
				regionName: liveBucketRegion,
				s3Client:   liveS3Client,
				accessKey:  accessKey,
				endpoint:   endpoint,
			},
		},
		BackupBucket: S3UnversionedBucket{
			S3Bucket{
				name:       backupBucketName,
				regionName: backupBucketRegion,
				s3Client:   backupS3Client,
				accessKey:  accessKey,
				endpoint:   endpoint,
			},
		},
	}, nil
}

func (p S3BucketPair) Backup(backupLocation string) (BackupBucketAddress, error) {
	files, err := p.LiveBucket.ListFiles()
	if err != nil {
		return BackupBucketAddress{}, err
	}
	for _, file := range files {
		err = p.BackupBucket.Copy(file, backupLocation, p.LiveBucket.Name(), p.LiveBucket.RegionName())
		if err != nil {
			return BackupBucketAddress{}, err
		}
	}
	return BackupBucketAddress{
		BucketName:   p.BackupBucket.Name(),
		BucketRegion: p.BackupBucket.RegionName(),
		Path:         backupLocation,
	}, nil
}
