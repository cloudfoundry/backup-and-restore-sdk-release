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

supported-stemcells=\
  ubuntu-bionic \
  ubuntu-jammy  \
  ubuntu-xenial \

supported-mariadb=\
  ubuntu-bionic~~~10.9-jammy  \
  ubuntu-jammy~~~~10.9-jammy  \
  ubuntu-xenial~~~10.9-jammy  \
  ubuntu-bionic~~~10.7-focal  \
  ubuntu-jammy~~~~10.7-focal  \
  ubuntu-xenial~~~10.7-focal  \
  ubuntu-bionic~~~10.5-focal  \
  ubuntu-jammy~~~~10.5-focal  \
  ubuntu-xenial~~~10.5-focal  \
  ubuntu-bionic~~~10.2-bionic \
  ubuntu-jammy~~~~10.2-bionic \
  ubuntu-xenial~~~10.2-bionic \

supported-mysql=\
  ubuntu-bionic~~~8.0-debian  \
  ubuntu-jammy~~~~8.0-debian  \
  ubuntu-xenial~~~8.0-debian  \
  ubuntu-bionic~~~8.0-oracle  \
  ubuntu-jammy~~~~8.0-oracle  \
  ubuntu-xenial~~~8.0-oracle  \
  ubuntu-bionic~~~5.7-debian  \
  ubuntu-jammy~~~~5.7-debian  \
  ubuntu-xenial~~~5.7-debian  \

supported-postgres=\
  ubuntu-bionic~~~13-bullseye  \
  ubuntu-jammy~~~~13-bullseye  \
  ubuntu-xenial~~~13-bullseye  \
  ubuntu-bionic~~~11-bullseye  \
  ubuntu-jammy~~~~11-bullseye  \
  ubuntu-xenial~~~11-bullseye  \
  ubuntu-bionic~~~10-bullseye  \
  ubuntu-jammy~~~~10-bullseye  \
  ubuntu-xenial~~~10-bullseye  \
  ubuntu-bionic~~~9.6-bullseye \
  ubuntu-jammy~~~~9.6-bullseye \
  ubuntu-xenial~~~9.6-bullseye \


docker-clean:
	docker-compose down --rmi local --volumes --remove-orphans ## remove containers created to run the tests

docker-clean-prune: $(supported-stemcells) ## remove containers AND IMAGES created to run the tests
docker-compile-release-matrix: $(supported-stemcells) ## run compilation test for all supported stemcells

docker-compile-release: ## run compilation test for the specified $STEMCELL_NAME
	@echo "\033[92mCompiling: Stemcell ${STEMCELL_NAME} \033[0m"
	docker-compose --log-level ERROR run dockerize-release

$(supported-stemcells):
	@if [ "$(MAKECMDGOALS)" = "docker-clean-prune" ]; then\
		export STEMCELL_NAME=$@                                  ;\
		docker-compose down --rmi all --volumes --remove-orphans ;\
	fi
	@if [ "$(MAKECMDGOALS)" = "docker-compile-release-matrix" ]; then\
		export STEMCELL_NAME=$@        ;\
		$(MAKE) docker-compile-release ;\
	fi

docker-system-mariadb: ## run system tests for the specified $STEMCELL_NAME and $MARIADB_VERSION
	@echo "\033[92mTesting: Stemcell ${STEMCELL_NAME} | MariaDB ${MARIADB_VERSION} \033[0m"
	export MARIADB_PASSWORD="$$(head /dev/urandom | md5sum | cut -f1 -d" ")"  ;\
	docker-compose --log-level ERROR run system-db-mariadb

docker-system-mysql: ## run system tests for the specified $STEMCELL_NAME and $MYSQL_VERSION
	@echo "\033[92mTesting: Stemcell ${STEMCELL_NAME} | Mysql ${MYSQL_VERSION} \033[0m"
	export MYSQL_PASSWORD="$$(head /dev/urandom | md5sum | cut -f1 -d" ")"  ;\
	docker-compose --log-level ERROR run system-db-mysql

docker-system-postgres: ## run system tests for the specified $STEMCELL_NAME and $POSTGRES_VERSION
	@echo "\033[92mTesting: Stemcell ${STEMCELL_NAME} | Postgres ${POSTGRES_VERSION} \033[0m"

	export POSTGRES_PASSWORD="$$(head /dev/urandom | md5sum | cut -f1 -d" ")"  ;\
	export ENABLE_TLS="no"                                                     ;\
	docker-compose --log-level ERROR run system-db-postgres
	docker-compose down -v --remove-orphans --rmi local

	export POSTGRES_PASSWORD="$$(head /dev/urandom | md5sum | cut -f1 -d" ")"  ;\
	export ENABLE_TLS="yes"                                                    ;\
	docker-compose --log-level ERROR run system-db-postgres
	docker-compose down -v --remove-orphans --rmi local

	export POSTGRES_PASSWORD="$$(head /dev/urandom | md5sum | cut -f1 -d" ")"  ;\
	export ENABLE_TLS="mutual"                                                 ;\
	docker-compose --log-level ERROR run system-db-postgres
	docker-compose down -v --remove-orphans --rmi local

docker-system-mariadb-matrix: $(supported-mariadb) ## run system tests for all supported Stemcells and MARIADB versions
$(supported-mariadb):
	export MARIADB_VERSION=$(word 2,$(subst ~, ,$@)) ;\
	export STEMCELL_NAME=$(word 1,$(subst ~, ,$@))   ;\
	$(MAKE) docker-system-mariadb

docker-system-mysql-matrix: $(supported-mysql) ## run system tests for all supported Stemcells and MYSQL versions
$(supported-mysql):
	export MYSQL_VERSION=$(word 2,$(subst ~, ,$@)) ;\
	export STEMCELL_NAME=$(word 1,$(subst ~, ,$@))   ;\
	$(MAKE) docker-system-mysql

docker-system-postgres-matrix: $(supported-postgres) ## run system tests for all supported Stemcells and POSTGRES versions
$(supported-postgres):
	export POSTGRES_VERSION=$(word 2,$(subst ~, ,$@)) ;\
	export STEMCELL_NAME=$(word 1,$(subst ~, ,$@))   ;\
	$(MAKE) docker-system-postgres

docker-system-matrix: docker-system-mariadb-matrix docker-system-mysql-matrix docker-system-postgres-matrix ## run all system tests for all supported Stemcells and database versions

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
system: docker-system ## run all system tests in Docker (same as docker-system-matrix)
clean: docker-clean ## remove containers created to run the tests (same as docker-clean)
