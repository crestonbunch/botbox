#!/bin/bash

go build -o api

docker build -t botbox-api:$BOTBOX_VERSION \
    --build-arg domain_name=$BOTBOX_DOMAIN_NAME \
    --build-arg db_host=botbox-database \
    --build-arg db_user=botbox \
    --build-arg db_password=$BOTBOX_DB_PASSWORD \
    --build-arg db_name=botbox \
    --build-arg db_sslmode=disable \
    --build-arg smtp_host=$BOTBOX_SMTP_HOST \
    --build-arg smtp_port=$BOTBOX_SMTP_PORT \
    --build-arg smtp_username=$BOTBOX_SMTP_USERNAME \
    --build-arg smtp_password=$BOTBOX_SMTP_PASSWORD \
    --build-arg recaptcha_secret=$BOTBOX_RECAPTCHA_SECRET \
    --build-arg recaptcha_sitekey=$BOTBOX_RECAPTCHA_SITEKEY \
    .
