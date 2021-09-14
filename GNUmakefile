TEST?=./...
PKG_NAME=pkg/artifactory
VERSION := $(shell git tag --sort=-creatordate | head -1 | sed  -n 's/v\([0-9]*\).\([0-9]*\).\([0-9]*\)/\1.\2.\3/p')
NEXT_VERSION := $(shell echo ${VERSION}| awk -F '.' '{print $$1 "." $$2 "." $$3 +1 }' )

default: build

install:
	mkdir -p terraform.d/plugins/registry.terraform.io/jfrog/artifactory/${NEXT_VERSION}/darwin_amd64 && \
		(test -f terraform-provider-artifactory || go build -ldflags="-X 'artifactory.Version=${NEXT_VERSION}'") && \
		mv terraform-provider-artifactory terraform.d/plugins/registry.terraform.io/jfrog/artifactory/${NEXT_VERSION}/darwin_amd64 && \
		terraform init

clean:
	rm -fR .terraform.d/ .terraform terraform.tfstate*

release:
	@git tag ${NEXT_VERSION} && git push --mirror
	@echo "Pushed ${NEXT_VERSION}"

build: fmtcheck
	go build -ldflags="-X 'artifactory.Version=${NEXT_VERSION}'"

debug_install:
	mkdir -p terraform.d/plugins/registry.terraform.io/jfrog/artifactory/${NEXT_VERSION}/darwin_amd64 && \
		(test -f terraform-provider-artifactory || go build -gcflags "all=-N -l" -ldflags="-X 'artifactory.Version=${NEXT_VERSION}-develop'") && \
		mv terraform-provider-artifactory terraform.d/plugins/registry.terraform.io/jfrog/artifactory/${NEXT_VERSION}/darwin_amd64 && \
		terraform init


test:
	@echo "==> Starting unit tests"
	go test $(TEST) -timeout=30s -parallel=4

attach:
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient attach $$(pgrep terraform-provider-artifactory)

acceptance: fmtcheck
	export TF_ACC=1
	test -n ARTIFACTORY_USERNAME && test -n ARTIFACTORY_PASSWORD && test -n ARTIFACTORY_URL \
		&& go test -v -parallel 1 ./pkg/...


fmt:
	@echo "==> Fixing source code with gofmt..."
	@gofmt -s -w ./$(PKG_NAME)
	(command -v goimports &> /dev/null || go get golang.org/x/tools/cmd/goimports) && goimports -w pkg/artifactory

fmtcheck:
	@echo "==> Checking that code complies with gofmt requirements..."
	@sh -c "find . -name '*.go' -not -name '*vendor*' -print0 | xargs -0 gofmt -l -s"

.PHONY: build fmt
