FROM golang:1.18-alpine3.15 AS builder

WORKDIR /go/src/terralist

# Install gcc
RUN apk add build-base

COPY go.mod go.sum ./

RUN go mod download

ADD cmd/terralist ./cmd/terralist/
ADD pkg ./pkg
ADD internal ./internal/

ARG VERSION="dev"
ARG COMMIT_HASH="n/a"
ARG BUILD_TIMESTAMP="n/a"

RUN go build -a -v -o terralist \
    -ldflags="\
      -X 'main.Version=${VERSION}' \
      -X 'main.CommitHash=${COMMIT_HASH}' \
      -X 'main.BuildTimestamp=${BUILD_TIMESTAMP}' \
      -X 'main.Mode=release'" \
    ./cmd/terralist/main.go

FROM alpine:3.16.2

COPY --from=builder /go/src/terralist/terralist /usr/local/bin

WORKDIR /root

ENTRYPOINT [ "terralist" ]
CMD [ "server" ]