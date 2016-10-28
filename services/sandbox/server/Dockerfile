FROM alpine:edge

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

# TODO: compile/copy go binary
WORKDIR /opt/go/src/github.com/crestonbunch/botbox/services/sandbox/server

CMD /bin/bash go run main.go
