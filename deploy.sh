#!/bin/bash

# Replace with the current version
export BOTBOX_VERSION=0.01

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Deploying Botbox version $BOTBOX_VERSION"

# Step 1. Ask for information
echo -n "Domain name (e.g. example.com): "
read domain
echo
export BOTBOX_DOMAIN_NAME=$domain

echo -n "SMTP host (e.g. smtp.gmail.com): "
read smtp_host
echo
export BOTBOX_SMTP_HOST=$smtp_host

echo -n "SMTP port (e.g. 587): "
read smtp_port
echo
export BOTBOX_SMTP_PORT=$smtp_port

echo -n "SMTP username: "
read smtp_user
echo
export BOTBOX_SMTP_USERNAME=$smtp_user

echo -n "SMTP password: "
read -s smtp_pass
echo
export BOTBOX_SMTP_PASSWORD=$smtp_pass

echo -n "Database password: "
read -s password
echo
export BOTBOX_DB_PASSWORD=$password

# Step 2. Build the database
echo "Setting up database service"
cd $DIR/services/database
docker stop botbox-database
docker rm botbox-database
./build.sh
./run.sh

echo "Setting up API service"
# Step 3. Start the API service
cd $DIR/services/api/server
docker stop botbox-api
docker rm botbox-api
./build.sh
./run.sh
