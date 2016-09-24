FROM ubuntu:latest

RUN apt-get update
RUN apt-get install -y git

# Install all supported language runtimes
RUN apt-get install -y golang-go
RUN apt-get install -y ruby
RUN apt-get install -y nodejs

# setup a go path
RUN mkdir -p /opt/go
ENV GOPATH /opt/go

# install dependencies
RUN go get golang.org/x/net/websocket

# copy the botbox source into gopath
RUN mkdir -p /opt/go/src/botbox
RUN git clone https://github.com/crestonbunch/botbox /opt/go/src/botbox

# Don't run things as root
RUN useradd -ms /bin/bash sandbox

# create a directory to drop files into
RUN mkdir /botbox-server

# copy the default run script -- this can be overwritten by the server
COPY run.sh /botbox-server/

RUN chown -R sandbox:sandbox /botbox-server

#USER sandbox
WORKDIR /botbox-server/
