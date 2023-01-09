default: build

build: fmt
	go build -o ./bin/tf-v5-migrator main.go

fmt:
	@echo "==> Fixing source code with gofmt..."
	@go fmt ./...

.PHONY: build fmt
