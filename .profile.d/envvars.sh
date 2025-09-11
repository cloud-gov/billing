#!/usr/bin/env bash
set -eo pipefail
# Extract postgres PG* environment variables from the VCAP_SERVICES variable provided by Cloud Foundry.
# CF buildpack apps source all scripts in .profile.d before running the app start command.

aws_rds_credentials=$(echo $VCAP_SERVICES | jq '."aws-rds"[0].credentials')

export PGDATABASE=$(echo $aws_rds_credentials | jq -r '.db_name')
export PGHOST=$(echo $aws_rds_credentials | jq -r '.host')
export PGPORT=$(echo $aws_rds_credentials | jq -r '.port')
export PGUSER=$(echo $aws_rds_credentials | jq -r '.username')
export PGPASSWORD=$(echo $aws_rds_credentials | jq -r '.password')
