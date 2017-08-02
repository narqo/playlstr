GO := go
GOFLAGS :=

GIT := git
DOCKER := docker

GIT_REV := $(shell $(GIT) describe --always --tags --dirty=-dev)

PKG := ./playlstr
OUTPUT_DIR := $(CURDIR)

BUILD.go = $(GO) build $(GOFLAGS)
TEST.go  = $(GO) test $(GOFLAGS)

go_packages := $(shell $(GO) list ./... | grep -v /vendor/)

.PHONY: all
all: build test

.PHONY: build
build:
	$(BUILD.go) -ldflags "-X main.revision=$(GIT_REV)" -o $(OUTPUT_DIR)/playlstr.out $(PKG)

.PHONY: test
test:
	$(TEST.go) -v $(go_packages)

.PHONY: vet
vet:
	$(GO) vet $(go_packages)

.PHONY: clean
clean:
	$(GO) clean $(GOFLAGS) -i .
	$(RM) $(OUTPUT_DIR)/playlstr.out
