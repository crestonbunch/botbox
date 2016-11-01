#!/bin/bash
docker run --name botbox-api \
    -d -p 8081:8081 \
    --restart "unless-stopped" \
    --link botbox-database \
    botbox-api:$BOTBOX_VERSION
