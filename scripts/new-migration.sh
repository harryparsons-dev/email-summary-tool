#!/bin/sh

set -eu

name=${1:-}

if [ -z "$name" ]; then
	printf 'Usage: make migration NAME=migration_name\n' >&2
	exit 2
fi

case "$name" in
	*[!a-z0-9_]* | [!a-z]* | *_)
		printf 'Migration name must start with a lowercase letter and contain only lowercase letters, numbers, and underscores.\n' >&2
		exit 2
		;;
esac

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
project_dir=$(dirname -- "$script_dir")
migrations_dir="$project_dir/database/migrations/sql"
timestamp=$(date -u '+%Y%m%d%H%M%S')
prefix="$migrations_dir/${timestamp}_${name}"
up_file="${prefix}.up.sql"
down_file="${prefix}.down.sql"

if [ -e "$up_file" ] || [ -e "$down_file" ]; then
	printf 'Migration version %s already exists; wait one second and try again.\n' "$timestamp" >&2
	exit 1
fi

touch "$up_file" "$down_file"

printf 'Created:\n  %s\n  %s\n' \
	"${up_file#"$project_dir"/}" \
	"${down_file#"$project_dir"/}"
