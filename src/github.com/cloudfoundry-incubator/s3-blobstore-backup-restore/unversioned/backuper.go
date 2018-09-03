package unversioned

import "fmt"

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
	err := b.checkPairsValidity()
	if err != nil {
		return err
	}

	err = b.checkCrossPairValidity()
	if err != nil {
		return err
	}

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

func (b Backuper) checkPairsValidity() error {
	for bucketId, bucketPair := range b.bucketPairs {
		err := bucketPair.CheckValidity()
		if err != nil {
			return fmt.Errorf("failed to backup bucket '%s': %s", bucketId, err.Error())
		}
	}
	return nil
}

func (b Backuper) checkCrossPairValidity() error {
	for bucketID, bucketPair := range b.bucketPairs {
		liveBucketName := bucketPair.LiveBucketName()
		for compareBucketID, compareBucketPair := range b.bucketPairs {
			if liveBucketName == compareBucketPair.BackupBucketName() {
				return fmt.Errorf("'%s' backup bucket can not be the same as '%s' live bucket", compareBucketID, bucketID)
			}
		}
	}

	return nil
}
