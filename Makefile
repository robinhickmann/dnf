.PHONY: *

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

all: fmt build run

build:
	go build -ldflags "-s -w" -trimpath -o bin/$(GOOS)-$(GOARCH)/dnf ./cmd

run:
	./bin/$(GOOS)-$(GOARCH)/dnf 

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
