FROM alpine:edge

# Install all supported language runtimes
RUN apk add --update --no-cache \
    musl-dev \
    bash \
    git \
    go \
    python3

# Install the Python SDK
RUN mkdir /tmp/botbox && \
    git clone https://github.com/crestonbunch/botbox /tmp/botbox
WORKDIR /tmp/botbox/games/tron/sdk/python
RUN python3 setup.py install

# Don't run things as root
RUN adduser -S sandbox

# create a directory to drop files into
RUN mkdir /botbox-client

# copy the default run script -- this can be overwritten by the user
COPY run.sh /botbox-client/

RUN chown -R root:root /botbox-client && \
    chmod 555 /botbox-client/run.sh

USER sandbox
WORKDIR /botbox-client/

CMD /bin/bash run.sh
