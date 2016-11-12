#!/bin/bash

go build -o api
docker build -t botbox-api:$BOTBOX_VERSION \
    --build-arg db_pass=$BOTBOX_DB_PASSWORD \
    --build-arg db_user=botbox \
    --build-arg db_name=botbox \
    --build-arg db_host=botbox-database \
    --build-arg smtp_identity=$BOTBOX_SMTP_IDENTITY \
    --build-arg smtp_username=$BOTBOX_SMTP_USERNAME \
    --build-arg smtp_password=$BOTBOX_SMTP_PASSWORD \
    --build-arg smtp_host=$BOTBOX_SMTP_HOST \
    --build-arg smtp_port=$BOTBOX_SMTP_PORT \
    --build-arg domain=$BOTBOX_DOMAIN_NAME \
    --build-arg recaptcha_secret=$BOTBOX_RECAPTCHA_SECRET \
    .
