.PHONY: build run test clean fmt vet

BINARY_NAME=snippets
BUILD_DIR=build

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/snippets

run:
	go run ./cmd/snippets

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)

fmt:
	go fmt ./...

vet:
	go vet ./...