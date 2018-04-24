package unversioned

//go:generate counterfeiter -o fakes/fake_clock.go . Clock
type Clock interface {
	Now() string
}

type Backuper struct {
	bucketPairs         map[string]BucketPair
	clock               Clock
	destinationArtifact Artifact
}

func NewBackuper(
	bucketPairs map[string]BucketPair,
	destinationArtifact Artifact,
	clock Clock,
) Backuper {
	return Backuper{
		bucketPairs:         bucketPairs,
		destinationArtifact: destinationArtifact,
		clock:               clock,
	}
}

func (b Backuper) Run() error {
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
