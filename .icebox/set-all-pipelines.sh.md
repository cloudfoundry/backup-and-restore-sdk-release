```bash
#!/usr/bin/env bash
set -euox pipefail
```

## Determining ```project root folder``` according to git
This is useful to avoid:
- Complex and fragile relative paths that break after moving things around.
- Absolute paths that break when run in a different environment.

```bash
PROJECT_ROOT=$(git rev-parse --show-toplevel)
```

## Parameterizing ```Concourse``` instance
This is useful to:
- Deploy to different environments running the same scripts.
- Reuse variables and reduce duplication across scripts.

```bash
CONCOURSE_URL="http://concourse:8080"
CONCOURSE_TEAM="main"
```

## Explicitly running each of the pipeline scripts
In most cases this process involve two steps:
- Running [lit](https://github.com/vijithassar/lit) to transform the literate Markdown into a runnable script.
- Sourcing lit's output, instead of running in a new shell.
Otherwise above variables wouldn't be preserved.

Listing each script explicitly instead of using a wildcard
- Improves legibility
- Increases traceability
- Ensures intentionality
- Redudces accidental executions

```bash
source $(lit --input "./set-pipeline/bbr-sdk-test-prs.sh.md")
```