FROM ubuntu:latest

RUN apt-get -y update && \
    apt-get -y upgrade && \
    apt-get -y install ca-certificates

ARG domain_name
ARG db_host
ARG db_user
ARG db_password
ARG db_name
ARG db_sslmode
ARG smtp_host
ARG smtp_port
ARG smtp_username
ARG smtp_password
ARG recaptcha_secret
ARG recaptcha_sitekey

ENV BOTBOX_DOMAIN_NAME=$domain_name
ENV BOTBOX_DB_HOST=$db_host
ENV BOTBOX_DB_USER=$db_user
ENV BOTBOX_DB_PASSWORD=$db_password
ENV BOTBOX_DB_NAME=$db_name
ENV BOTBOX_DB_SSLMODE=$db_sslmode
ENV BOTBOX_SMTP_HOST=$smtp_host
ENV BOTBOX_SMTP_PORT=$smtp_port
ENV BOTBOX_SMTP_USERNAME=$smtp_username
ENV BOTBOX_SMTP_PASSWORD=$smtp_password
ENV BOTBOX_RECAPTCHA_SECRET=$recaptcha_secret
ENV BOTBOX_RECAPTCHA_SITEKEY=$recaptcha_sitekey

EXPOSE 8081

COPY api /bin/botbox/
COPY emails/* /bin/botbox/emails/

WORKDIR /bin/botbox
ENTRYPOINT /bin/botbox/api
