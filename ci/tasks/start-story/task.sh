#!/bin/bash

# Script that starts the story found with the find-pr-story task

set -euo pipefail

API_TOKEN=${TRACKER_API_TOKEN}
PROJECT_ID=${TRACKER_PROJECT_ID}
STORY_ID=$(cat ./tracker-story/id)

current_state=$(
    curl \
        -X GET \
        -H "X-TrackerToken: $API_TOKEN" \
        -H "Content-Type: application/json" \
        "https://www.pivotaltracker.com/services/v5/projects/$PROJECT_ID/stories/$STORY_ID" \
        -s \
        | jq '.current_state'
    )

if [[ "$current_state" != "unstarted" ]]; then
    echo "Story with ID: $STORY_ID has already been started, refusing to reset it's state to 'started'"
    exit 0
fi

err=$(
    curl \
        -X PUT \
        -H "X-TrackerToken: $API_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"current_state": "started","estimate": 1}' \
        "https://www.pivotaltracker.com/services/v5/projects/$PROJECT_ID/stories/$STORY_ID" \
        -s \
        | jq '.error'
    )

if [ "$err" != "null" ]
then
    echo "API call failed:" $err
    exit 1
fi

echo "Started story with ID: $STORY_ID"
