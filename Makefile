BIN_DIR  = bin
BINARY   = $(BIN_DIR)/wt

.PHONY: build test run clean tidy

build:
	go build -o $(BINARY) ./cmd/wt

test:
	go test ./...

run: build
	./$(BINARY)

clean:
	rm -rf $(BIN_DIR)

tidy:
	go mod tidy
