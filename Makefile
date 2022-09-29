.PHONY: bump-postgres

# To populate private.yml with S3 creds needed to upload blobs
config/private.yml:
	lpass show "Shared-PCF Backup and Restore/private_yml" --notes > config/private.yml

bump-postgres: config/private.yml
	./scripts/bump_postgres_blobs.bash

docker-unit-blobstore-azure:
	docker-compose run unit-blobstore-azure

docker-unit-blobstore-gcs:
	docker-compose run unit-blobstore-gcs

docker-unit-blobstore-s3:
	docker-compose run unit-blobstore-s3

docker-unit-database:
	docker-compose run unit-database

docker-unit-template-specs:
	docker-compose run unit-sdk-template

docker-unit: docker-unit-blobstore-azure docker-unit-blobstore-gcs docker-unit-blobstore-s3 docker-unit-database docker-unit-template-specs

local-unit-blobstore-azure:
	./src/azure-blobstore-backup-restore/scripts/run-unit-tests.bash

local-unit-blobstore-gcs:
	./src/gcs-blobstore-backup-restore/scripts/run-unit-tests.bash

local-unit-blobstore-s3:
	./src/s3-blobstore-backup-restore/scripts/run-unit-tests.bash

local-unit-database:
	./src/database-backup-restore/scripts/run-unit-tests.bash

local-unit-template-specs:
	./scripts/run-template-tests.bash

local-unit: local-unit-blobstore-azure local-unit-blobstore-gcs local-unit-blobstore-s3 local-unit-database local-unit-template-specs

unit: docker-unit
