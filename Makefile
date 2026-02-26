.PHONY: build run test clean help

BIN_DIR := bin
BINARY := $(BIN_DIR)/api

help:
	@echo "Available targets:"
	@echo "  make build    - Build the API binary"
	@echo "  make run      - Build and run the API"
	@echo "  make test     - Run all tests"
	@echo "  make clean    - Remove binaries"

build: $(BIN_DIR)
	go build -o $(BINARY) ./cmd/api

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

run: build
	./$(BINARY)

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)
