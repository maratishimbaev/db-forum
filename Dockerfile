FROM golang:latest AS build

# Copy source code to docker container
ADD . /opt/build/app

# Set workdir
WORKDIR /opt/build/app

# Build package
RUN go build cmd/api/main.go

FROM ubuntu:18.04 AS release

# Install postgres
ENV PGVERSION 10
RUN apt -y update && apt install -y postgresql-$PGVERSION

# Enter as user postgres
USER postgres

# Set env vars
ENV PGPASSWORD 'postgres'
ENV FORUM_USER 'postgres'
ENV FORUM_PASSWORD 'postgres'
ENV FORUM_DBNAME 'dbname'

# Set workdir
WORKDIR /opt/build/app

# Copy server
COPY --from=build /opt/build/app .

# Create postgres role and database
RUN /etc/init.d/postgresql start && \
    psql --command "ALTER USER postgres WITH SUPERUSER PASSWORD 'postgres';" && \
    createdb -E utf8 -T template0 -O postgres dbname && \
    psql dbname postgres -h localhost -f db.sql && \
    /etc/init.d/postgresql stop

# Configurate postgres
RUN echo "host all all 0.0.0.0/0 md5" >> /etc/postgresql/$PGVERSION/main/pg_hba.conf

# Add listen_address to .conf file
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVERSION/main/postgresql.conf

# Expose postgres port
EXPOSE 5432

# Add volumes
VOLUME ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

# Expose app port
EXPOSE 5000

# Run postgres and app
CMD service postgresql start && ./main
