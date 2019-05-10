.PHONY: bump-postgres

# To populate private.yml with S3 creds needed to upload blobs
config/private.yml:
	lpass show "Shared-PCF Backup and Restore/private_yml" --notes > config/private.yml

bump-postgres: config/private.yml
	./scripts/bump_postgres_blobs.bash

