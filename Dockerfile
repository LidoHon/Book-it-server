# Use official PostgreSQL image
FROM postgres:15

# Install pg_cron extension
RUN apt-get update && \
    apt-get install -y postgresql-15-cron && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copy any custom initialization scripts (optional)
COPY ./init.sql /docker-entrypoint-initdb.d/
