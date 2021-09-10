#!/bin/bash

source in.bash

clean () {
    unset -f cd
    unset -f wget
}

T_get_download_url_generates_url_for_version() {
    local output="$(get_download_url '8.0.25')"
    local expected='https://downloads.mysql.com/archives/get/p/23/file/mysql-8.0.25.tar.gz'

    if [[ "$output" != "$expected" ]]
    then
        $T_fail "expected $expected got $output"
        return 0
    fi
}

T_build_output_returns_json() {
    local output=$(build_output '8.0.25')
    local expected='{"version": {"ref": "8.0.25"}}'

    if [[ "$output" != "$expected" ]]
    then
        $T_fail "expected $expected got $output"
        return 0
    fi
}

T_download_file_calls_wget_with_correct_url() {
    wget() {
        if [[ "$1" != 'https://downloads.mysql.com/archives/get/p/23/file/mysql-8.0.25.tar.gz' ]]
        then
            $T_fail "wget called with wrong url: $1"
            return 0
        fi
    }

    cd() {
        return 0
    }

    download_file 'https://downloads.mysql.com/archives/get/p/23/file/mysql-8.0.25.tar.gz' 'test_assets'

    clean
}

T_download_file_cds_to_directory() {
    cd() {
        if [[ "$1" != 'test_assets' ]]
        then
            $T_fail "cd called with wrong directory: $1"
            return 0
        fi
    }

    wget() {
        return 0
    }

    download_file 'https://downloads.mysql.com/archives/get/p/23/file/mysql-8.0.25.tar.gz' 'test_assets'

    clean
}

T_in_script() {
    local input=$(cat test_assets/test_input.json)

    wget() {
        if [[ "$1" != 'https://downloads.mysql.com/archives/get/p/23/file/mysql-5.6.45.tar.gz' ]]
        then
            $T_fail "wget called with wrong url: $1"
            return 0
        fi
    }
    export -f wget

    cd() {
        if [[ "$1" != 'test_assets' ]]
        then
            $T_fail "cd called with wrong directory: $1"
            return 0
        fi
    }
    export -f cd

    local output=$(echo $input | ./in test_assets)
    local expected='{"version": {"ref": "5.6.45"}}'

    clean

    if [[ "$output" != "$expected" ]]
    then
        $T_fail "expected $expected got $output"
        return 0
    fi
}