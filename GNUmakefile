TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=pkg/artifactory

default: build

build: fmtcheck
	go install

test: fmtcheck
	go test $(TEST) -timeout=30s -parallel=4

artifactory:
	@echo "==> Launching Artifactory in Docker..."
	@scripts/run-artifactory.sh

docker:
	@docker build -t dillongiacoppo/terraform-artifactory .

testacc: fmtcheck artifactory
	TF_ACC=1 ARTIFACTORY_USERNAME=admin ARTIFACTORY_PASSWORD=password ARTIFACTORY_URL=http://localhost:8080/artifactory \
	go test $(TEST) -v -parallel 20 $(TESTARGS) -timeout 120m
	@docker stop artifactory

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w ./$(PKG_NAME)
	goimports -w pkg/artifactory

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

.PHONY: build test testacc fmt