FROM postgres:9.6.1

ARG root_pw
ENV POSTGRES_PASSWORD=$root_pw
ENV POSTGRES_USER=botbox
ENV POSTGRES_DB=botbox
ENV POSTGRES_INITDB_ARGS="--encoding=UTF8"
ENV PGDATA="/var/lib/postgresql/data/botbox"

COPY initdb.d/* /docker-entrypoint-initdb.d/
