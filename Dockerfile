FROM node:18-alpine3.19 AS frontend

WORKDIR /home/node/terralist

COPY ./web/package.json ./web/yarn.lock ./
RUN yarn install --frozen-lockfile

ARG VERSION="dev"
ENV TERRALIST_VERSION=${VERSION}

COPY ./web ./
RUN yarn build

FROM golang:1.22-alpine3.19 AS backend

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

FROM alpine:3.17

RUN addgroup terralist && \
    adduser -S -G terralist terralist && \
    adduser terralist root && \
    chown terralist:root /home/terralist/ && \
    chmod g=u /home/terralist/ && \
    chmod g=u /etc/passwd

RUN apk add --no-cache \
      git~=2.38 \
      libcap~=2.66 \
      dumb-init~=1.2 \
      su-exec~=0.2

COPY docker-entrypoint.sh /usr/local/bin/
COPY --from=backend /go/src/terralist/terralist /usr/local/bin

ENTRYPOINT [ "docker-entrypoint.sh" ]
CMD [ "server" ]
