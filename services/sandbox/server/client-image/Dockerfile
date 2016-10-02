FROM ubuntu:latest

# Install all supported language runtimes
RUN apt-get -y update && apt-get -y upgrade && apt-get -y install && \
apt-get install -y \
    git \
    golang-go \
    ruby \
    nodejs \
    python3 python3-pip python3-setuptools

# Install the Python SDK
RUN mkdir /tmp/botbox && \
    git clone https://github.com/crestonbunch/botbox /tmp/botbox
WORKDIR /tmp/botbox/games/tron/sdk/python
RUN python3 setup.py install

# Don't run things as root
RUN useradd -ms /bin/bash sandbox

# create a directory to drop files into
RUN mkdir /botbox-client

# copy the default run script -- this can be overwritten by the user
COPY run.sh /botbox-client/

RUN chown -R sandbox:sandbox /botbox-client && \
    chmod 555 /botbox-client/run.sh

#USER sandbox
WORKDIR /botbox-client/

CMD /bin/bash run.sh
