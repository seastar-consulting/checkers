VERSION ?= $(shell git describe --tags --always --dirty)
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
BINARY_NAME=checkers
BINARY_DIR=bin

# Python configuration
PYTHON_ROOT=/opt/homebrew/opt/python@3.13/Frameworks/Python.framework/Versions/3.13
PYTHON_CONFIG=$(PYTHON_ROOT)/lib/pkgconfig
export PKG_CONFIG_PATH=$(PYTHON_CONFIG)

# Build platforms
PLATFORMS=linux darwin windows
ARCHITECTURES=amd64 arm64

# Version flags
VERSION_FLAGS=-X github.com/seastar-consulting/checkers/internal/version.Version=$(VERSION) \
              -X github.com/seastar-consulting/checkers/internal/version.GitCommit=$(GIT_COMMIT) \
              -X github.com/seastar-consulting/checkers/internal/version.GitBranch=$(GIT_BRANCH)

# Test version flags - use fixed values for predictable test output
TEST_VERSION=v1.2.3-test
TEST_VERSION_FLAGS=-X github.com/seastar-consulting/checkers/internal/version.Version=$(TEST_VERSION)

# Build targets
.PHONY: all
all: build

.PHONY: build
build:
	mkdir -p $(BINARY_DIR)
	go build -v -ldflags="$(VERSION_FLAGS)" -o $(BINARY_DIR)/$(BINARY_NAME)

.PHONY: release
release:
	@mkdir -p $(BINARY_DIR)
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			output_name=$(BINARY_NAME)-$$platform-$$arch; \
			if [ $$platform = "windows" ]; then \
				output_name=$$output_name.exe; \
			fi; \
			echo "Building $$platform/$$arch..."; \
			GOOS=$$platform GOARCH=$$arch go build -v -ldflags="$(VERSION_FLAGS)" -o $(BINARY_DIR)/$$output_name; \
		done; \
	done

.PHONY: test
test:
	go test -v -ldflags="$(TEST_VERSION_FLAGS)" ./...

.PHONY: clean
clean:
	rm -rf $(BINARY_DIR)
