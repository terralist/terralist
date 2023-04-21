FROM node:19-alpine3.15 AS frontend

WORKDIR /home/node/terralist

COPY ./web/package.json ./web/yarn.lock ./
RUN yarn install --frozen-lockfile

COPY ./web ./
RUN yarn build

FROM golang:1.19-alpine3.15 AS backend

WORKDIR /go/src/terralist

# Install gcc
RUN apk add build-base

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/terralist/ ./cmd/terralist
COPY pkg/ ./pkg/
COPY internal/ ./internal/
COPY web/ ./web/
COPY --from=frontend /home/node/terralist/dist ./web/dist

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

COPY --from=backend /go/src/terralist/terralist /usr/local/bin

WORKDIR /root

ENTRYPOINT [ "terralist" ]
CMD [ "server" ]