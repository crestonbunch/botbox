#!/bin/bash

docker run -d -p 8081:8081 \
    --restart "unless-stopped" \
    --name botbox-api \
    --link botbox-database \
    botbox-api:$BOTBOX_VERSION \
