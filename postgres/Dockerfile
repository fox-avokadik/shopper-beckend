FROM postgres:latest

RUN apt-get update && apt-get install -y postgresql-cron

RUN echo "shared_preload_libraries = 'pg_cron'" >> /etc/postgresql/postgresql.conf

COPY init.sql /docker-entrypoint-initdb.d/

CMD ["postgres"]
