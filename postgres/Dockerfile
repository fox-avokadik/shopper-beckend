FROM postgres:16.8

RUN apt-get update && \
    apt-get install -y postgresql-16-cron && \
    rm -rf /var/lib/apt/lists/*

CMD ["postgres"]