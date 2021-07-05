.PHONY: bump-postgres

# To populate private.yml with S3 creds needed to upload blobs
config/private.yml:
	lpass show "Shared-PCF Backup and Restore/private_yml" --notes > config/private.yml

bump-postgres: config/private.yml
	./scripts/bump_postgres_blobs.bash

specs:
	bundle
	bundle exec rspec

unit-db:
	cd ./src/database-backup-restore && ginkgo -r -v -skipPackage system_tests

unit-s3:
	cd ./src/s3-blobstore-backup-restore && ginkgo -mod vendor -r -keepGoing -p --skipPackage s3bucket

unit-azure:
	cd ./src/azure-blobstore-backup-restore && ginkgo -mod vendor -r -keepGoing -p --skipPackage contract_test

unit-gcs:
	cd ./src/gcs-blobstore-backup-restore && ginkgo -mod vendor -r -keepGoing -p --skipPackage contract_test

unit: specs unit-db unit-s3 unit-azure unit-gcs

