FROM debian:stable-slim

RUN apt-get update && \
    apt-get -y install ca-certificates

ADD gopath/bin/tm /usr/bin/tm
