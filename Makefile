.PHONY: all build run fmt clean lint tls

VERSION := $(shell git describe --tags --always 2>/dev/null || echo dev)
BUILD_TIME := $(shell date +"%Y-%m-%dT%H:%M:%S%z")

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
BIN := bin/$(GOOS)-$(GOARCH)/dnf

all: tls fmt lint build run

build:
	go build -ldflags "-s -w -X main.version=$(VERSION) \
		-X main.buildTime=$(BUILD_TIME)" -trimpath -o $(BIN) ./cmd/dnf

run:
	./$(BIN)

fmt:
	go fmt ./...

clean:
	rm -rf bin

lint: ./bin/golangci-lint
	./bin/golangci-lint run

./bin/golangci-lint:
	curl -sSfL https://golangci-lint.run/install.sh | sh -s v2.12.2

tls: key.pem cert.pem

key.pem cert.pem:
	openssl req -x509 \
  		-newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
  		-keyout key.pem -out cert.pem \
  		-days 365 -nodes \
  		-subj "/CN=localhost"
