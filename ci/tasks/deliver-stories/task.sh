#!/bin/bash

# deliver-stories
# Script that looks for finished tracker stories related to PRs from a specific git repo and sets them as delivered

set -euo pipefail

API_TOKEN=${TRACKER_API_TOKEN}
PROJECT_ID=${TRACKER_PROJECT_ID}
GIT_REPOSITORY=${GIT_REPOSITORY}

echo "Fetching accepted tracker stories for $GIT_REPOSITORY PRs..."

res=$(
    curl \
        -H "X-TrackerToken: $API_TOKEN" \
        "https://www.pivotaltracker.com/services/v5/projects/$PROJECT_ID/stories?with_label=github-pull-request&with_state=finished" \
        -s \
        | jq -r '.[] | select(.description | contains("https://github.com/'$GIT_REPOSITORY'")) | .id'
    )

if [ -z "$res" ]
then
    echo "No stories found"
    exit 0
fi

echo $res

readarray -t ids <<<"$res"

for id in "${ids[@]}"
do
    err=$(
        curl \
            -X PUT \
            -H "X-TrackerToken: $API_TOKEN" \
            -H "Content-Type: application/json" \
            -d '{"current_state": "delivered"}' \
            "https://www.pivotaltracker.com/services/v5/projects/$PROJECT_ID/stories/$id" \
            -s \
            | jq '.error'
        )

    if [ "$err" != "null" ]
    then
        echo "API call failed:" $err
        exit 1
    fi
done

echo "Delivered ${#ids[@]} PR stories"