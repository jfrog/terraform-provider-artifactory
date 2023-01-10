default: build

build: fmt
	@echo "Building the binary..."
	go build -o ./bin/v5-v6-migrator main.go

fmt:
	@echo "Fixing source code with 'go fmt'..."
	@go fmt ./...

.PHONY: build fmt
