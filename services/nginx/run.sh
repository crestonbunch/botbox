#/bin/bash

docker run --name botbox-nginx \
    -p 80:80 \
    --link botbox-web \
    --link botbox-api \
    -d botbox-nginx:$BOTBOX_VERSION
