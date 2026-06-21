.PHONY: *

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_TIME := $(shell date +"%Y-%m-%dT%H:%M:%S%z")

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

all: fmt build run

build:
	go build -ldflags "-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)" \
		-trimpath -o bin/$(GOOS)-$(GOARCH)/dnf ./cmd

run:
	./bin/$(GOOS)-$(GOARCH)/dnf --version

fmt:
	go fmt ./...

clean:
	rm -rf bin

tls:
	openssl req -x509 \
  		-newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
  		-keyout key.pem -out cert.pem \
  		-days 365 -nodes \
  		-subj "/CN=localhost"
