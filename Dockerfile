FROM golang:1.18 AS build

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /tm

COPY go.mod go.sum /tm/
RUN go mod download

ARG GIT_TAG
ENV GIT_TAG=${GIT_TAG:-unknown}

COPY . .
RUN make install && \
    rm -rf ${GOPATH}/src && \
    rm -rf ${HOME}/.cache

FROM debian:stable-slim

RUN apt-get update && \
    apt-get install --no-install-recommends --no-install-suggests -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /go/bin/tm /bin/tm
