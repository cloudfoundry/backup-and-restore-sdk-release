###### Help ###################################################################

$(VERBOSE).SILENT:
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
  ubuntu-bionic~~~5.7-debian  \
  ubuntu-jammy~~~~5.7-debian  \
  ubuntu-xenial~~~5.7-debian  \
  ubuntu-bionic~~~8.0-debian  \
  ubuntu-jammy~~~~8.0-debian  \
  ubuntu-xenial~~~8.0-debian  \
  ubuntu-bionic~~~8.0-oracle  \
  ubuntu-jammy~~~~8.0-oracle  \
  ubuntu-xenial~~~8.0-oracle  \

supported-postgres=\
  ubuntu-bionic~~~15-bullseye  \
  ubuntu-jammy~~~~15-bullseye  \
  ubuntu-xenial~~~15-bullseye  \
  ubuntu-bionic~~~13-bullseye  \
  ubuntu-jammy~~~~13-bullseye  \
  ubuntu-xenial~~~13-bullseye  \
  ubuntu-bionic~~~11-bullseye  \
  ubuntu-jammy~~~~11-bullseye  \
  ubuntu-xenial~~~11-bullseye  \


docker-clean: ## remove containers created to run the tests
	if ! echo "$@" | grep -q "${FOCUS}" ; then                                  \
		echo "\033[92mSkipping $@ \033[0m"                                 ;\
	else                                                                        \
		docker --log-level ERROR compose down --rmi local --volumes --remove-orphans ;\
	fi

docker-clean-prune: $(supported-stemcells) ## remove containers AND IMAGES created to run the tests
docker-system-mariadb: $(supported-mariadb) ## run system tests for all supported Stemcells and MARIADB versions
docker-system-mysql: $(supported-mysql) ## run system tests for all supported Stemcells and MYSQL versions
docker-system-postgres: $(supported-postgres) ## run system tests for all supported Stemcells and POSTGRES versions

docker-system: docker-system-postgres docker-system-mariadb docker-system-mysql ## run all system tests for all supported Stemcells and database versions

docker-unit-blobstore-azure: ## run azure blobstore unit tests in Docker
	docker compose up --build --exit-code-from unit-blobstore-azure unit-blobstore-azure

docker-unit-blobstore-gcs: ## run GCS blobstore unit tests in Docker
	docker compose up --build --exit-code-from unit-blobstore-gcs unit-blobstore-gcs

docker-unit-blobstore-s3: ## run S3 blobstore unit tests in Docker
	docker compose up --build --exit-code-from unit-blobstore-s3 unit-blobstore-s3

docker-unit-database: ## run database unit tests in Docker
	docker compose up --build --exit-code-from unit-database unit-database

docker-unit-template-specs: ## run templating unit tests in Docker
	docker compose up --build --exit-code-from unit-sdk-template unit-sdk-template

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
system: docker-system ## run all system tests in Docker (same as docker-system)
clean: docker-clean ## remove containers created to run the tests (same as docker-clean)

$(supported-stemcells):
	if ! echo "$@" | grep -q "${FOCUS}" ; then                                  \
		echo "\033[92mSkipping $@ \033[0m"                                 ;\
	else                                                                        \
		if [ "$(MAKECMDGOALS)" = "docker-clean-prune" ]; then               \
			echo "\033[92mCleaning $@ \033[0m"                         ;\
			export STEMCELL_NAME=$@                                    ;\
			docker --log-level ERROR compose down --rmi all --volumes --remove-orphans   ;\
		fi                                                                  \
	fi

$(supported-mariadb):
	if ! echo "$@" | grep -q "${FOCUS}" ; then                                  \
		echo "\033[92mSkipping $@ \033[0m"                                 ;\
	else                                                                        \
		echo "\033[92m Testing: MariaDB $@ \033[0m"                        ;\
		export MARIADB_VERSION=$(word 2,$(subst ~, ,$@))                   ;\
		export STEMCELL_NAME=$(word 1,$(subst ~, ,$@))                     ;\
		export MARIADB_PASSWORD="$$(head /dev/urandom | md5sum | cut -f1 -d" ")"  ;\
		docker --log-level ERROR compose up --build --exit-code-from system-db-mariadb system-db-mariadb         ;\
		exit_code=$$?                                                      ;\
		docker --log-level ERROR compose rm --stop --force -v system-db-mariadb-backing-db ;\
		if [ $${exit_code} -ne 0 ]; then exit 1; fi                        ;\
	fi

$(supported-mysql):
	if ! echo "$@" | grep -q "${FOCUS}" ; then                                  \
		echo "\033[92mSkipping $@ \033[0m"                                 ;\
	else                                                                        \
		echo "\033[92m Testing MySQL $@ \033[0m"                           ;\
		export MYSQL_VERSION=$(word 2,$(subst ~, ,$@))                     ;\
		export STEMCELL_NAME=$(word 1,$(subst ~, ,$@))                     ;\
		export MYSQL_PASSWORD="$$(head /dev/urandom | md5sum | cut -f1 -d" ")"    ;\
		docker --log-level ERROR compose up --build --exit-code-from system-db-mysql system-db-mysql          ;\
		exit_code=$$?                                                      ;\
		docker --log-level ERROR compose rm --stop --force -v system-db-mysql-backing-db ;\
		if [ $${exit_code} -ne 0 ]; then exit 1; fi                        ;\
	fi

docker-system-postgres-aux:
	export POSTGRES_VERSION=$(word 2,$(subst ~, ,$(MATRIX_TUPLE)))             ;\
	export STEMCELL_NAME=$(word 1,$(subst ~, ,$(MATRIX_TUPLE)))                ;\
	export POSTGRES_PASSWORD="$$(head /dev/urandom | md5sum | cut -f1 -d" ")"  ;\
	docker --log-level ERROR compose up --build --exit-code-from system-db-postgres system-db-postgres          ;\
	exit_code=$$?                                                              ;\
	docker --log-level ERROR compose rm --stop --force -v system-db-postgres-backing-db  ;\
	if [ $${exit_code} -ne 0 ]; then exit 1; fi                                ;\

$(supported-postgres):
	if ! echo "$@" | grep -q "${FOCUS}" ; then                                  \
		echo "\033[92mSkipping $@ \033[0m"                                 ;\
	else                                                                        \
		echo "\033[92m Testing Postgres $@ \033[0m"                        ;\
		export ENABLE_TLS="no"                                             ;\
		$(MAKE) docker-system-postgres-aux MATRIX_TUPLE=$@                 ;\
		if [ $$? -ne 0 ]; then exit 1; fi                                  ;\
		export ENABLE_TLS="yes"                                            ;\
		$(MAKE) docker-system-postgres-aux MATRIX_TUPLE=$@                 ;\
		if [ $$? -ne 0 ]; then exit 1; fi                                  ;\
		export ENABLE_TLS="mutual"                                         ;\
		$(MAKE) docker-system-postgres-aux MATRIX_TUPLE=$@                 ;\
		if [ $$? -ne 0 ]; then exit 1; fi                                  ;\
	fi
