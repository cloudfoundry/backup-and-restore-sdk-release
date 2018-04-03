package blobstore

//go:generate counterfeiter -o fakes/fake_clock.go . Clock
type Clock interface {
	Now() string
}

type UnversionedBackuper struct {
	bucketPairs         map[string]UnversionedBucketPair
	clock               Clock
	destinationArtifact UnversionedArtifact
}

func (b UnversionedBackuper) Run() error {
	timestamp := b.clock.Now()
	addresses := map[string]BackupBucketAddress{}
	for id, pair := range b.bucketPairs {
		address, err := pair.Backup(timestamp + "/" + id)
		if err != nil {
			return err
		}
		addresses[id] = address
	}
	return b.destinationArtifact.Save(addresses)
}

func NewUnversionedBackuper(
	bucketPairs map[string]UnversionedBucketPair,
	destinationArtifact UnversionedArtifact,
	clock Clock,
) UnversionedBackuper {
	return UnversionedBackuper{
		bucketPairs:         bucketPairs,
		destinationArtifact: destinationArtifact,
		clock:               clock,
	}
}
