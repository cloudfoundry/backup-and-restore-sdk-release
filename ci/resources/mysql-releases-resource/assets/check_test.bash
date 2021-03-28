#!/bin/bash

source check.bash

T_get_latest_release_returns_empty_string_on_emtpy_input() {
    local output=$(get_latest_release 'MySQL Community Server 5.6')

    if [[ -n "$output" ]]
    then
        $T_fail "get_latest_release should have failed on empty pipe input, got $output"
        return 0
    fi
}

T_get_latest_release_returns_version_number() {
    local xml=$(cat test_assets/test_rss.xml)
    local output="$(echo $xml | get_latest_release 'MySQL Community Server 5.6')"
    local expected='5.6.51'

    if [[ "$output" != "$expected" ]]
    then
        $T_fail "expected $expected got $output"
        return 0
    fi
}

T_build_output_returns_json() {
    local output=$(build_output '5.6.51')
    local expected='[{"ref": "5.6.51"}]'

    if [[ "$output" != "$expected" ]]
    then
        $T_fail "expected $expected got $output"
        return 0
    fi
}

T_check_script() {
    curl() {
        echo $(cat test_assets/test_rss.xml)
    }
    export -f curl

    local input=$(cat test_assets/test_input.json)
    local output=$(echo $input | ./check)
    local expected='[{"ref": "5.6.51"}]'

    if [[ "$output" != "$expected" ]]
    then
        $T_fail "expected $expected got $output"
        return 0
    fi
}

T_check_script_returns_empty_string_on_empty_input() {
    curl() {
        echo $(cat test_assets/test_rss.xml)
    }
    export -f curl

    local output=$(./check)
    local expected=

    if [[ "$output" != "$expected" ]]
    then
        $T_fail "expected $expected got $output"
        return 0
    fi
}