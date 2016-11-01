#!/bin/bash
docker run --name botbox-database \
    -v botbox-volume:/var/lib/postgresql/data \
    -d -p 5432:5432 \
    --restart "unless-stopped" \
    botbox-database:$BOTBOX_VERSION
