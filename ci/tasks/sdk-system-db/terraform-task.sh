#!/usr/bin/env bash

# Copyright (C) 2019-Present Pivotal Software, Inc. All rights reserved.
#
# This program and the accompanying materials are made available under
# the terms of the under the Apache License, Version 2.0 (the "License”);
# you may not use this file except in compliance with the License.
#
# You may obtain a copy of the License at
# http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#
# See the License for the specific language governing permissions and
# limitations under the License.

set -eu

terraform_output() {
  terraform output -state=terraform-state/terraform.tfstate "$1" | jq -r .
}

if [ "$DB_TYPE" == "mysql" ]; then
  export MYSQL_HOSTNAME="$( terraform_output "${DB_PREFIX}_address" )"

  if [ ! -z "$MYSQL_CA_CERT_PATH" ]; then
    export MYSQL_CA_CERT="$( cat "cert-store/${MYSQL_CA_CERT_PATH}" )"
  fi

  if [ ! -z "$MYSQL_CLIENT_CERT_PATH" ]; then
    export MYSQL_CLIENT_CERT="$( cat "cert-store/${MYSQL_CLIENT_CERT_PATH}" )"
  fi

  if [ ! -z "$MYSQL_CLIENT_KEY_PATH" ]; then
    export MYSQL_CLIENT_KEY="$( cat "cert-store/${MYSQL_CLIENT_KEY_PATH}" )"
  fi

elif [ "$DB_TYPE" == "postgres" ]; then
  export POSTGRES_HOSTNAME="$( terraform_output "${DB_PREFIX}_address" )"

  if [ ! -z "$POSTGRES_CA_CERT_PATH" ]; then
    export POSTGRES_CA_CERT="$( cat "cert-store/${POSTGRES_CA_CERT_PATH}" )"
  fi

  if [ ! -z "$POSTGRES_CLIENT_CERT_PATH" ]; then
    export POSTGRES_CLIENT_CERT="$( cat "cert-store/${POSTGRES_CLIENT_CERT_PATH}" )"
  fi

  if [ ! -z "$POSTGRES_CLIENT_KEY_PATH" ]; then
    export POSTGRES_CLIENT_KEY="$( cat "cert-store/${POSTGRES_CLIENT_KEY_PATH}" )"
  fi

else
  >&2 echo "Invalid DB_TYPE, please use mysql or postgres!"
  exit 1
fi

backup-and-restore-ci/tasks/sdk-system-db/task.sh

