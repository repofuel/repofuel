FROM golang:1.18-buster AS builder

WORKDIR /go/src

COPY . .

ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux

RUN go build -o /http-server ./accounts/cmd/http-server \
	&& rm -rf "${GOPATH}/pkg/mod"



FROM debian:buster-slim

RUN apt-get update && apt-get install -y \
	ca-certificates \
	&& update-ca-certificates --fresh \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /http-server .

CMD ./http-server
