FROM golang:1.18-buster AS builder

WORKDIR /go/src

ENV LIBGIT2_VER "1.1.0"
ENV GIT2GO_MAJOR_VER "v31"

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
    libssl-dev \
    libssh2-1-dev \
    && rm -rf /var/lib/apt/lists/*

COPY . .

ENV GO111MODULE on
ENV CGO_ENABLED 1
ENV GOOS linux

RUN go get -v -t -d ./ingest/... \
    && sh ./ingest/script/build-libgit2.sh --static --system \
    && go build -o /http-server --tags "static,system_libgit2" ./ingest/cmd/http-server \
    && rm -rf "${GOPATH}/pkg/mod"



FROM debian:buster-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && update-ca-certificates --fresh \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /http-server .

CMD ./http-server
