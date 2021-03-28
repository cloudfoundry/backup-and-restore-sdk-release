#!/bin/bash

set -e

basht ./check_test.bash
basht ./in_test.bash

echo "All tests passed"