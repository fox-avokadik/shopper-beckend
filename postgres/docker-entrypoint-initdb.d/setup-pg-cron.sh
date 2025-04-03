#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    ALTER SYSTEM SET shared_preload_libraries = 'pg_cron';
    ALTER SYSTEM SET cron.database_name = 'testdb';
EOSQL

pg_ctl -D /var/lib/postgresql/data restart