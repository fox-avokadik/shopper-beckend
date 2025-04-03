FROM postgres:16

RUN apt-get update && \
    apt-get install -y postgresql-16-cron && \
    rm -rf /var/lib/apt/lists/*

COPY <<EOF /docker-entrypoint-initdb.d/setup-pg-cron.sh
#!/bin/bash
set -e

echo "shared_preload_libraries = 'pg_cron'" >> /var/lib/postgresql/data/postgresql.conf
echo "cron.database_name = 'testdb'" >> /var/lib/postgresql/data/postgresql.conf

# Restart PostgreSQL to apply changes
pg_ctl -D /var/lib/postgresql/data restart
EOF

RUN chmod +x /docker-entrypoint-initdb.d/setup-pg-cron.sh

CMD ["postgres"]