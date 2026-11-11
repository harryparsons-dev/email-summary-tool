#!/bin/sh
# Creates the role the API connects as. It owns the application database, so it
# can run migrations, but it is not a superuser: it cannot read server files,
# run programs on the database host, or touch other databases.
#
# PostgreSQL runs this once, when the data volume is first initialised.
set -eu

: "${DB_USER:?DB_USER must be set}"
: "${DB_PASSWORD:?DB_PASSWORD must be set}"

psql -v ON_ERROR_STOP=1 \
	--username "$POSTGRES_USER" \
	--dbname "$POSTGRES_DB" \
	-v app_user="$DB_USER" \
	-v app_password="$DB_PASSWORD" \
	-v app_db="$POSTGRES_DB" <<'EOSQL'
CREATE ROLE :"app_user" LOGIN PASSWORD :'app_password';
ALTER DATABASE :"app_db" OWNER TO :"app_user";
EOSQL
