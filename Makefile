# VERSION := $(shell git describe --tags --abbrev=0)
# REVISION := $(shell git rev-parse --short HEAD)
# LDFLAGS := -X 'main.version=$(VERSION)' \
# 		   -X 'main.revision=$(REVISION)'
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -X 'main.version=$(REVISION)'

# 必要なツール類をセットアップする
## Setup
setup:
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/goimports
	go get github.com/songmu/make2help/cmd/make2help

# go modを使って依存パッケージをインストールする
## Install dependencies
deps: setup
	go mod download

# Test
test: deps
	go test ./...

# Lint
lint: deps
	golint ./...

# Format source codes
fmt: deps
	goimports -w ./...

# Run go-slack-ansible
run:
	go run cmd/go-slack-ansible/*.go

# Build binaries ex. make bin/go-slack-ansible
bin/%: cmd/% deps
	CGO_ENABLED=0 go build -ldflags "-s -w" -ldflags "$(LDFLAGS)" -o $@ $</*.go

# Show help
help:
	@make2help $(MAKEFILE_LIST)

.PHONY: deps test lint help
