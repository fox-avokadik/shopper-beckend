FROM postgres:latest

RUN apt-get update && apt-get install -y \
    postgresql-server-dev-all \
    build-essential \
    curl \
    git

RUN git clone --depth 1 https://github.com/citusdata/pg_cron.git && \
    cd pg_cron && \
    make && \
    make install

RUN echo "shared_preload_libraries = 'pg_cron'" >> /etc/postgresql/postgresql.conf

CMD ["postgres"]