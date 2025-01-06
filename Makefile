BINARY_NAME=checkers
BINARY_DIR=bin

.PHONY: build
build:
	mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$(BINARY_NAME)

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	rm -rf $(BINARY_DIR)
