SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
.DEFAULT_GOAL := all

define INFO
	@echo -e "\e[32m----- $(1)\e[0m"
endef

all: build
.PHONY: all

################################################################################
# Dependencies
vendor: go.mod go.sum
	@$(call INFO,"Tidying Go modules")
	go mod tidy
	@$(call INFO,"Vendoring Go modules")
	go mod vendor

tools/go-licenses:
	@$(call INFO,"Installing tool $(shell basename $@)")
	GOBIN=$$(realpath $$(dirname $@)) go install github.com/google/go-licenses@v1.6.0

tools/golangci-lint:
	@$(call INFO,"Installing tool $(shell basename $@)")
	GOBIN=$$(realpath $$(dirname $@)) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2

################################################################################
# Linters (and checks)
lint: lint-go check-licenses
.PHONY: lint

lint-go: vendor tools/golangci-lint
	@$(call INFO,"Linting Go")
	./tools/golangci-lint run -v
.PHONY: lint-go

check-licenses: tools/go-licenses vendor
	@$(call INFO,"Checking third-party licenses")
	./tools/go-licenses check ./... --include_tests --disallowed_types unknown,forbidden,restricted
.PHONY: check-licenses

################################################################################
# Builds
build: bin/fleeting-plugin-proxmox
.PHONY: build

bin/fleeting-plugin-proxmox: vendor $(shell find cmd -name *.go)
	@$(call INFO,"Building $@")
	@mkdir -p $(shell dirname $@)
	go build -o $@ ./cmd/fleeting-plugin-proxmox

################################################################################
# Tests
test: unit-test integration-test
.PHONY: test

unit-test:
	@$(call INFO,"Running unit tests")
	go test -v ./cmd/...
.PHONY: unit-test

integration-test: bin/fleeting-plugin-proxmox
	@$(call INFO,"Running integration tests")
	go test -v $(shell go list ./test/integration) \
		-plugin-binary-path="$(PWD)/bin/fleeting-plugin-proxmox" \
		-config-path="$(PWD)/config.json"
.PHONY: integration-test

################################################################################
# Cleanup
clean:
	@$(call INFO,"Cleaning")
	rm -rf \
		bin \
		tools/go-licenses \
		tools/golangci-lint \
		vendor
.PHONY: clean
