package gcs

type BackupAction struct{}

func NewBackupAction() BackupAction {
	return BackupAction{}
}

func (b BackupAction) Run(backuper Backuper, artifact BackupArtifact) error {
	err := backuper.CreateLiveBucketSnapshot()
	if err != nil {
		return err
	}

	_, err = backuper.TransferBlobsToBackupBucket()
	if err != nil {
		return err
	}

	backupBucketAddresses := map[string]BackupBucketAddress{}
	err = backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses)
	if err != nil {
		return err
	}

	err = artifact.Write(backupBucketAddresses)
	if err != nil {
		return err
	}

	return nil
}
