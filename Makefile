# File: Makefile

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
VERSION := 1.0.0

default: build

.PHONY: build
build:
	go build -o terraform-provider-hlb_v$(VERSION)

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/$(HOSTNAME)/hlb/$(VERSION)/$(GOOS)_$(GOARCH)
	mv terraform-provider-hlb_v$(VERSION) ~/.terraform.d/plugins/$(HOSTNAME)/hlb/$(VERSION)/$(GOOS)_$(GOARCH)/

.PHONY: test
test:
	go test -v ./...

.PHONY: testacc
testacc:
	TF_ACC=1 go test -v ./... -timeout 120m

.PHONY: clean
clean:
	rm -f terraform-provider-hlb_v$(VERSION)