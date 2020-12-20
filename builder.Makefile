include vars.mk

NULL := /dev/null

ifeq ($(COMMIT),)
  COMMIT := $(shell git rev-parse --short HEAD 2> $(NULL))
endif

ifeq ($(TAG_NAME),)
  TAG_NAME := $(shell git describe --always --dirty --abbrev=10 2> $(NULL))
endif

PKG_NAME=github.com/glours/docker-lint
STATIC_FLAGS= CGO_ENABLED=0
LDFLAGS := "-s -w \
  -X $(PKG_NAME)/internal.GitCommit=$(COMMIT) \
  -X $(PKG_NAME)/internal.Version=$(TAG_NAME)"
GO_BUILD = $(STATIC_FLAGS) go build -trimpath -ldflags=$(LDFLAGS)

ifneq ($(strip $(E2E_TEST_NAME)),)
	RUN_TEST=-test.run $(E2E_TEST_NAME)
endif

VARS:= DOCKER_CONFIG=$(PWD)/docker-config

.PHONY: lint
lint:
	golangci-lint run --timeout 10m0s ./...

.PHONY: e2e
e2e:
	mkdir -p docker-config/cli-plugins
	cp ./bin/${PLATFORM_BINARY} docker-config/cli-plugins/${BINARY}
	# TODO: gotestsum doesn't forward ldflags to go test with golang 1.15.0, so moving back to go test temporarily
	$(VARS) go test ./e2e $(RUN_TEST) -ldflags=$(LDFLAGS)

.PHONY: test-unit
test-unit:
	gotestsum $(shell go list ./... | grep -vE '/e2e')

cross:
	GOOS=linux   GOARCH=amd64 $(GO_BUILD) -o dist/docker-lint_linux_amd64 ./cmd/docker-lint
	GOOS=darwin  GOARCH=amd64 $(GO_BUILD) -o dist/docker-lint_darwin_amd64 ./cmd/docker-lint
	GOOS=windows GOARCH=amd64 $(GO_BUILD) -o dist/docker-lint_windows_amd64.exe ./cmd/docker-lint

.PHONY: build
build:
	mkdir -p bin
	$(GO_BUILD) -o bin/$(PLATFORM_BINARY) ./cmd/docker-lint
