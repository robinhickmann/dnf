.PHONY: all build run dev release $(PLATFORMS) install uninstall reinstall fmt clean lint tls

VERSION := $(shell git describe --tags --always 2>/dev/null || echo dev)
BUILD_TIME := $(shell date +"%Y-%m-%dT%H:%M:%S%z")

# List of supported platforms built for release
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64 windows-arm64

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
BIN := bin/$(GOOS)-$(GOARCH)/dnf

INSTALL_BIN := /usr/local/bin/dnf
INSTALL_SERVICE := /etc/systemd/system/dnf.service
INSTALL_DIR := /etc/dnf

all: tls fmt lint build run

build:
	go build -ldflags "-s -w -X main.version=$(VERSION) \
		-X main.buildTime=$(BUILD_TIME)" -trimpath -o $(BIN) ./cmd/dnf

run:
	@./$(BIN) --config config.dev.yml

dev:
	@docker compose up --build --quiet-build

release: $(PLATFORMS)

$(PLATFORMS):
	$(MAKE) build \
		GOOS=$(word 1,$(subst -, ,$@)) \
		GOARCH=$(word 2,$(subst -, ,$@)) \

install: build
	@[ "$(shell id -u)" -eq 0 ] || (echo "error: run with sudo"; exit 1)

	install -Dm755 $(BIN) $(INSTALL_BIN)
	install -Dm644 dnf.service $(INSTALL_SERVICE)
	install -Dm644 config.yml $(INSTALL_DIR)/config.yml

	systemctl daemon-reload
	systemctl enable --now dnf

uninstall:
	@[ "$(shell id -u)" -eq 0 ] || (echo "error: run with sudo"; exit 1)
	
	systemctl disable --now dnf
	rm -rf $(INSTALL_BIN) $(INSTALL_SERVICE) $(INSTALL_DIR)
	systemctl daemon-reload

reinstall: uninstall install

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
