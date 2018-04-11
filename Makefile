GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
PROJDIR := $(realpath $(CURDIR))
SOURCEDIR := $(PROJDIR)/
BUILDDIR := $(PROJDIR)/dist

all: test build

dep:
	@misc/scripts/deps-ensure
	@dep ensure -v

fmt:
	@gofmt -w $(GOFMT_FILES)

test:
	@go test -v ./...
	@go vet ./...

.PHONY: clean build
clean:
	-@rm -rf ./dist/
