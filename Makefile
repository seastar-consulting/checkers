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
	@mkdir -p $(BINARY_DIR)
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			output_name=$(BINARY_NAME)-$$platform-$$arch; \
			if [ $$platform = "windows" ]; then \
				output_name=$$output_name.exe; \
			fi; \
			echo "Building $$platform/$$arch..."; \
			GOOS=$$platform GOARCH=$$arch go build -v -ldflags="-X main.Version=$(VERSION)" -o $(BINARY_DIR)/$$output_name; \
		done; \
	done

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	rm -rf $(BINARY_DIR)
