FROM ubuntu:latest

RUN apt-get update

# Install all supported language runtimes
RUN apt-get install -y golang-go
RUN apt-get install -y ruby
RUN apt-get install -y nodejs

# Don't run things as root
RUN useradd -ms /bin/bash sandbox

# create a directory to drop files into
RUN mkdir /botbox-client

# copy the default run script -- this can be overwritten by the user
COPY run.sh /botbox-client/

RUN chown -R sandbox:sandbox /botbox-client

#USER sandbox
WORKDIR /botbox-client/
