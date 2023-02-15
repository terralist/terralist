FROM golang:1.18-alpine3.15 AS builder

WORKDIR /go/src/terralist
# Install gcc
RUN apk add build-base

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/terralist ./cmd/terralist/
COPY pkg ./pkg
COPY internal ./internal/
COPY entrypoint.sh ./entrypoint.sh

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

FROM alpine:3.15

COPY --from=builder /go/src/terralist/terralist /usr/local/bin
COPY --from=builder /go/src/terralist/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

WORKDIR /root

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]