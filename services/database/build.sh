#!/bin/bash

docker volume create --name botbox-volume
docker build -t botbox-database:$BOTBOX_VERSION \
    --build-arg root_pw=$BOTBOX_DB_PASSWORD \
    .
