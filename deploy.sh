#!/bin/bash

# Replace with the current version
export BOTBOX_VERSION=0.01

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Deploying Botbox version $BOTBOX_VERSION"

# Step 1. Load environment variables
source env.sh

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

echo "Setting up web client"
# Step 4. Start the web client service
cd $DIR/services/web
docker stop botbox-web
docker rm botbox-web
./build.sh
./run.sh

echo "Setting up reverse proxy"
# Step 5. Start the nginx reverse proxy
cd $DIR/services/nginx
docker stop botbox-nginx
docker rm botbox-nginx
./build.sh
./run.sh
