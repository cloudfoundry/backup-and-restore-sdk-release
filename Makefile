###### Help ###################################################################

.DEFAULT_GOAL = help

.PHONY: help

help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m - %s\n", $$1, $$2}'

config/private.yml: # To populate private.yml with S3 creds needed to upload blobs
	lpass show "Shared-PCF Backup and Restore/private_yml" --notes > config/private.yml

.PHONY: bump-postgres
bump-postgres: config/private.yml ## update blobs, spec and packaging to PostgreSQL version specified by MAJOR and MINOR
	./scripts/bump_postgres_blobs.bash

docker-unit-blobstore-azure: ## run azure blobstore unit tests in Docker
	docker-compose run unit-blobstore-azure

docker-unit-blobstore-gcs: ## run GCS blobstore unit tests in Docker
	docker-compose run unit-blobstore-gcs

docker-unit-blobstore-s3: ## run S3 blobstore unit tests in Docker
	docker-compose run unit-blobstore-s3

docker-unit-database: ## run database unit tests in Docker
	docker-compose run unit-database

docker-unit-template-specs: ## run templating unit tests in Docker
	docker-compose run unit-sdk-template

docker-unit: docker-unit-blobstore-azure docker-unit-blobstore-gcs docker-unit-blobstore-s3 docker-unit-database docker-unit-template-specs ## run all unit tests in Docker

local-unit-blobstore-azure: ## run azure blobstore unit tests locally
	./src/azure-blobstore-backup-restore/scripts/run-unit-tests.bash

local-unit-blobstore-gcs: ## run GCS blobstore unit tests locally
	./src/gcs-blobstore-backup-restore/scripts/run-unit-tests.bash

local-unit-blobstore-s3: ## run S3 blobstore unit tests locally
	./src/s3-blobstore-backup-restore/scripts/run-unit-tests.bash

local-unit-database: ## run database unit tests locally
	./src/database-backup-restore/scripts/run-unit-tests.bash

local-unit-template-specs: ## run templating unit tests locally
	./scripts/run-template-tests.bash

local-unit: local-unit-blobstore-azure local-unit-blobstore-gcs local-unit-blobstore-s3 local-unit-database local-unit-template-specs ## run all unit tests locally

unit: docker-unit ## run all unit tests in Docker (same as docker-unit)
