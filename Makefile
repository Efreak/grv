VERSION=$(shell git describe --long --tags --dirty --always 2>/dev/null || echo 'Unknown')
HEAD_OID=$(shell git rev-parse --short HEAD 2>/dev/null || echo 'Unknown')
BUILD_DATETIME=$(shell date '+%Y-%m-%d %H:%M:%S %Z')

GOCMD=go
GOLINT=golint

BINARY?=grv
SOURCE_DIR=./cmd/grv
LDFLAGS=-X 'main.version=$(VERSION)' -X 'main.headOid=$(HEAD_OID)' -X 'main.buildDateTime=$(BUILD_DATETIME)'
STATIC_LDFLAGS=-extldflags '-lncurses -ltinfo -lgpm -static'
BUILD_FLAGS=--tags static -ldflags "$(LDFLAGS)"
STATIC_BUILD_FLAGS=--tags static -ldflags "$(LDFLAGS) $(STATIC_LDFLAGS)"

GRV_DIR:=$(dir $(realpath $(lastword $(MAKEFILE_LIST))))
GOPATH_DIR:=$(shell go env GOPATH)
GOBIN_DIR:=$(GOPATH_DIR)/bin
GIT2GO_DIR:=$(GOPATH_DIR)/src/gopkg.in/libgit2/git2go.v25
GIT2GO_PATCH=git2go.v25.patch

all: $(BINARY)

$(BINARY): build-libgit2
	$(GOCMD) build $(BUILD_FLAGS) -o $(BINARY) $(SOURCE_DIR)

.PHONY: install
install: $(BINARY)
	install -m755 -d $(GOBIN_DIR)
	install -m755 $(BINARY) $(GOBIN_DIR)

.PHONY: update
update:
	$(GOCMD) get -d ./...

.PHONY: update-test
update-test:
	$(GOCMD) get github.com/golang/lint/golint
	$(GOCMD) get github.com/stretchr/testify/mock

.PHONY: build-libgit2
build-libgit2: update
	if patch --dry-run -N -d $(GIT2GO_DIR) -p1 < $(GIT2GO_PATCH) >/dev/null; then \
		patch -d $(GIT2GO_DIR) -p1 < $(GIT2GO_PATCH); \
	fi
	cd $(GIT2GO_DIR) && git submodule update --init;
	make -C $(GIT2GO_DIR) install-static

# Only tested on Ubuntu.
# Requires dependencies static library versions to be present alongside dynamic ones
.PHONY: static
static: build-libgit2
	$(GOCMD) build $(STATIC_BUILD_FLAGS) -o $(BINARY) $(SOURCE_DIR)

.PHONY: test
test: $(BINARY) update-test
	$(GOCMD) test $(BUILD_FLAGS) $(SOURCE_DIR)
	$(GOCMD) vet $(SOURCE_DIR)
	$(GOLINT) -set_exit_status $(SOURCE_DIR)

.PHONY: clean
clean:
	rm -f $(BINARY)
