#!/bin/sh

set -a
. ./docker.env
set +a

cmd=migrate
mig="sql/migrations"
cfg="sql/migrations/tern.conf"

go tool tern "$cmd" -c "$cfg" -m "$mig" "$@"
