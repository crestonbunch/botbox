FROM alpine:edge

# Install all supported language runtimes
RUN apk add --update --no-cache \
    musl-dev \
    bash \
    git \
    go

# setup a go path
RUN mkdir -p /opt/go && export GOPATH=/opt/go
ENV GOPATH /opt/go

# install dependencies
RUN go get golang.org/x/net/websocket && \
    go get github.com/docker/engine-api && \
    go get github.com/docker/go-connections && \
    go get github.com/crestonbunch/botbox/...

# Don't run things as root
RUN adduser -S sandbox

# create a directory to drop files into
RUN mkdir /botbox-server

# copy the default run script -- this can be overwritten by the server
COPY run.sh /botbox-server/

RUN chown -R sandbox:root /botbox-server && \
    chmod 555 /botbox-server/run.sh

#USER sandbox
WORKDIR /botbox-server/

CMD /bin/bash run.sh
