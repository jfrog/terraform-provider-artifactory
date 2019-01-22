TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=pkg/artifactory

default: build

build: fmtcheck
	go install

test: fmtcheck
	go test $(TEST) -timeout=30s -parallel=4

docker:
	@echo "==> Launching Artifactory in Docker..."
	@scripts/run-artifactory.sh

testacc: fmtcheck docker
	TF_ACC=1 ARTIFACTORY_USERNAME=admin ARTIFACTORY_PASSWORD=password ARTIFACTORY_URL=http://localhost:8080/artifactory \
	go test $(TEST) -v -parallel 20 $(TESTARGS) -timeout 120m
	@docker stop artifactory

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w ./$(PKG_NAME)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

.PHONY: build test testacc fmt