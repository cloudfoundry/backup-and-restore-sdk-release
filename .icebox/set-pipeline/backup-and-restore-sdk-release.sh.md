```bash
#!/usr/bin/env bash
set -euox pipefail
echo $CONCOURSE_URL
```

## Setting Pipelines
```bash
fly --target=concourse                          \
    login                                       \
    --concourse-url=${CONCOURSE_URL}            \
    --team-name=${CONCOURSE_TEAM}

fly --target=concourse sync

fly --target=concourse                          \
    set-pipeline                                \
    --non-interactive                           \
    --pipeline=backup-and-restore-sdk-release   \
    --config="${PROJECT_ROOT}/ci/pipelines/backup-and-restore-sdk-release/pipeline.yml"
```