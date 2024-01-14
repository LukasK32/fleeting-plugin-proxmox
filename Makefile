SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
.DEFAULT_GOAL := all

all: test build
.PHONY: all

vendor: go.mod go.sum
	@go mod vendor

test: vendor
	@go test -v ./...
.PHONY: test

build: bin/fleeting-plugin-proxmox
.PHONY: build

bin/fleeting-plugin-proxmox: vendor $(shell find cmd -name *.go)
	@mkdir -p $(shell dirname $@)
	@go build -a -ldflags "-w -extldflags '-static'" -o $@ ./cmd/fleeting-plugin-proxmox

clean:
	@rm -rf vendor bin
.PHONY: clean
