TEST?=./...
PKG_NAME=pkg/xray
#VERSION := $(shell git tag --sort=-creatordate | head -1 | sed  -n 's/v\([0-9]*\).\([0-9]*\).\([0-9]*\)/\1.\2.\3/p')
# Replace explicit version after the first release
VERSION := 0.0.0
NEXT_VERSION := $(shell echo ${VERSION}| awk -F '.' '{print $$1 "." $$2 "." $$3 +1 }' )

default: build

install:
	mkdir -p terraform.d/plugins/registry.terraform.io/jfrog/xray/${NEXT_VERSION}/darwin_amd64 && \
		(test -f terraform-provider-xray || go build -ldflags="-X 'xray.Version=${NEXT_VERSION}'") && \
		mv terraform-provider-xray terraform.d/plugins/registry.terraform.io/jfrog/xray/${NEXT_VERSION}/darwin_amd64 && \
		terraform init

clean:
	rm -fR .terraform.d/ .terraform terraform.tfstate* terraform.d/

release:
	@git tag ${NEXT_VERSION} && git push --mirror
	@echo "Pushed ${NEXT_VERSION}"

build: fmtcheck
	go build -ldflags="-X 'xray.Version=${NEXT_VERSION}'"

debug_install:
	mkdir -p terraform.d/plugins/registry.terraform.io/jfrog/xray/${NEXT_VERSION}/darwin_amd64 && \
		(test -f terraform-provider-xray || go build -gcflags "all=-N -l" -ldflags="-X 'xray.Version=${NEXT_VERSION}-develop'") && \
		mv terraform-provider-xray terraform.d/plugins/registry.terraform.io/jfrog/xray/${NEXT_VERSION}/darwin_amd64 && \
		terraform init


test:
	@echo "==> Starting unit tests"
	go test $(TEST) -timeout=30s -parallel=4

attach:
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient attach $$(pgrep terraform-provider-xray)

acceptance: fmtcheck
	export TF_ACC=1
	test -n ARTIFACTORY_USERNAME && test -n ARTIFACTORY_URL && test -n ARTIFACTORY_ACCESS_TOKEN \
		&& go test -v -parallel 20 ./pkg/...


fmt:
	@echo "==> Fixing source code with gofmt..."
	@gofmt -s -w ./$(PKG_NAME)
	(command -v goimports &> /dev/null || go get golang.org/x/tools/cmd/goimports) && goimports -w pkg/xray

fmtcheck:
	@echo "==> Checking that code complies with gofmt requirements..."
	@sh -c "find . -name '*.go' -not -name '*vendor*' -print0 | xargs -0 gofmt -l -s"

.PHONY: build fmt
