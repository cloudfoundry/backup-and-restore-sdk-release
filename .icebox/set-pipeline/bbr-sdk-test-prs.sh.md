```bash
#!/usr/bin/env bash
set -euox pipefail
echo $CONCOURSE_URL
```

## Setting Pipelines
```bash
fly --target=concourse                      \
    login                                   \
    --concourse-url=${CONCOURSE_URL}        \
    --team-name=${CONCOURSE_TEAM}

fly --target=concourse sync

fly --target=concourse                      \
    set-pipeline                            \
    --non-interactive                       \
    --pipeline=bbr-sdk-test-prs             \
    --config="${PROJECT_ROOT}/ci/pipelines/bbr-sdk-test-prs/pipeline.yml"
```