#!/bin/bash

WEB_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
echo $WEB_DIR

docker run --name botbox-web \
    -d -p 8082:80 \
    -v $WEB_DIR/dist/:/usr/share/nginx/html:ro \
    --restart "unless-stopped" \
    botbox-web:$BOTBOX_VERSION
