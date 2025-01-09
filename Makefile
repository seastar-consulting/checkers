VERSION ?= $(shell git describe --tags --always --dirty)
BINARY_NAME=checkers
BINARY_DIR=bin

# Build platforms
PLATFORMS=linux darwin windows
ARCHITECTURES=amd64 arm64

# Build targets
.PHONY: all
all: build

.PHONY: build
build:
	mkdir -p $(BINARY_DIR)
	go build -v -ldflags="-X main.Version=$(VERSION)" -o $(BINARY_DIR)/$(BINARY_NAME)

.PHONY: release
release:
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			mkdir -p $(BINARY_DIR)/$$platform-$$arch; \
			GOOS=$$platform GOARCH=$$arch go build -v -ldflags="-X main.Version=$(VERSION)" -o $(BINARY_DIR)/$$platform-$$arch/$(BINARY_NAME)$(if $(findstring windows,$$platform),.exe,); \
		done; \
	done

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	rm -rf $(BINARY_DIR)
