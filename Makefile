VERSION := 1.0.0
HOSTNAME := registry.terraform.io
NAMESPACE := zonehero.io
NAME := hlb
BINARY := terraform-provider-${NAME}
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)
LOCAL_PROVIDER_DIR := ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS}_${ARCH}

.PHONY: build install clean

default: build

build:
	go build -o ${BINARY}_v${VERSION}

install: build
	@mkdir -p ${LOCAL_PROVIDER_DIR}
	mv ${BINARY}_v${VERSION} ${LOCAL_PROVIDER_DIR}/${BINARY}_v${VERSION}

clean:
	rm -f ${BINARY}_v${VERSION}
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

# Optional: Generate documentation
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

.PHONY: test test-race test-short fmt lint docs
