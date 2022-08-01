FROM golang:1.18-buster AS builder

WORKDIR /go/src

COPY . .

ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux

RUN	go build -o /http-server ./ml/cmd/http-server \
	&& rm -rf "${GOPATH}/pkg/mod"



FROM python:3.7-slim-buster

COPY ./ml/python /python
COPY --from=builder /http-server .

RUN apt-get update && apt-get install -y \
	ca-certificates \
	&& update-ca-certificates --fresh \
	&& rm -rf /var/lib/apt/lists/* \
	&& pip3 install --no-cache-dir -r /python/requirments.txt

CMD ./http-server
