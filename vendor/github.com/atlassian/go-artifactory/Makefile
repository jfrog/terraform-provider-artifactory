all: install-hooks test

install-hooks:
	@misc/scripts/install-hooks

dep:
	@misc/scripts/deps-ensure
	@dep ensure -v

fmt:
	@go fmt ./...

test:
	@go test -v -race ./...
