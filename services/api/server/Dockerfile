FROM ubuntu:latest

RUN apt-get -y update && \
    apt-get -y upgrade && \
    apt-get -y install ca-certificates

ARG db_pass
ARG db_user
ARG db_host
ARG db_name
ENV POSTGRES_DB_USER=$db_user
ENV POSTGRES_DB_NAME=$db_name
ENV POSTGRES_DB_PASSWORD=$db_pass
ENV POSTGRES_DB_HOST=$db_host

ARG smtp_identity
ARG smtp_username
ARG smtp_password
ARG smtp_host
ARG smtp_port
ENV SMTP_IDENTITY=$smtp_identity
ENV SMTP_USERNAME=$smtp_username
ENV SMTP_PASSWORD=$smtp_password
ENV SMTP_HOST=$smtp_host
ENV SMTP_PORT=$smtp_port

ARG domain
ENV BOTBOX_DOMAIN_NAME=$domain

ARG recaptcha_secret
ENV RECAPTCHA_SECRET=$recaptcha_secret

EXPOSE 8081
COPY api /bin/
CMD /bin/api
