VERSION := 1.0.0
HOSTNAME := registry.terraform.io
NAMESPACE := orographiclift
NAME := hlb
BINARY := terraform-provider-${NAME}
CLI_BINARY := zonehero
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)
LOCAL_PROVIDER_DIR := ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS}_${ARCH}
LOCAL_BIN_DIR := /usr/local/bin

.PHONY: build install clean cli cli-install

default: build

build:
	go build -o ${BINARY}_v${VERSION}

install: build
	@mkdir -p ${LOCAL_PROVIDER_DIR}
	mv ${BINARY}_v${VERSION} ${LOCAL_PROVIDER_DIR}/${BINARY}_v${VERSION}

cli:
	go build -o ${CLI_BINARY} ./cmd/zonehero

cli-install: cli
	@sudo mv ${CLI_BINARY} ${LOCAL_BIN_DIR}/${CLI_BINARY}
	@echo "Installed ${CLI_BINARY} to ${LOCAL_BIN_DIR}"

clean:
	rm -f ${BINARY}_v${VERSION}
	rm -f ${CLI_BINARY}
	rm -rf ${LOCAL_PROVIDER_DIR}

test:
	go test ./... -v

# Optional: Run tests with race condition checking
test-race:
	go test -race ./... -v

# Optional: Run only quick tests (useful for rapid development cycles)
test-short:
	go test ./... -v -short

# Optional: Check for formatting issues
fmt:
	go fmt ./...

# Optional: Run linter
lint:
	golangci-lint run

# Generate documentation using tfplugindocs
docs:
	go generate ./...

all: build cli

install-all: install cli-install

.PHONY: test test-race test-short fmt lint docs all install-all
