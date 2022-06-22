TEST?=./...
GO_ARCH=$(shell go env GOARCH)
TARGET_ARCH=$(shell go env GOOS)_${GO_ARCH}
ifeq ($(GO_ARCH), amd64)
GORELEASER_ARCH=${TARGET_ARCH}_$(shell go env GOAMD64)
endif
PKG_NAME=pkg/artifactory

# if this path ever changes, you need to also update the 'ldflags' value in .goreleaser.yml
PKG_VERSION_PATH=github.com/jfrog/terraform-provider-artifactory/v6/${PKG_NAME}/provider

VERSION := $(shell git tag --sort=-creatordate | head -1 | sed  -n 's/v\([0-9]*\).\([0-9]*\).\([0-9]*\)/\1.\2.\3/p')
NEXT_VERSION := $(shell echo ${VERSION}| awk -F '.' '{print $$1 "." $$2 "." $$3 +1 }' )
BUILD_PATH=terraform.d/plugins/registry.terraform.io/jfrog/artifactory/${NEXT_VERSION}/${TARGET_ARCH}

default: build

install:
	rm -fR .terraform.d && \
	mkdir -p ${BUILD_PATH} && \
		(test -f terraform-provider-artifactory || GORELEASER_CURRENT_TAG=${NEXT_VERSION} goreleaser build --single-target --rm-dist --snapshot) && \
		mv -v dist/terraform-provider-artifactory_${GORELEASER_ARCH}/terraform-provider-artifactory_v${NEXT_VERSION}* ${BUILD_PATH} && \
		rm -f .terraform.lock.hcl && \
		sed -i.bak 's/version = ".*"/version = "${NEXT_VERSION}"/' sample.tf && rm sample.tf.bak && \
		terraform init

clean:
	rm -fR terraform.d/ .terraform terraform.tfstate* terraform.d/ .terraform.lock.hcl

release:
	@git tag ${NEXT_VERSION} && git push --mirror
	@echo "Pushed ${NEXT_VERSION}"
	GOPROXY=https://proxy.golang.org GO111MODULE=on go get github.com/jfrog/terraform-provider-artifactory@v${NEXT_VERSION}
	@echo "Updated pkg cache"

update_pkg_cache:
	GOPROXY=https://proxy.golang.org GO111MODULE=on go get github.com/jfrog/terraform-provider-artifactory@v${VERSION}

build: fmtcheck
	GORELEASER_CURRENT_TAG=${NEXT_VERSION} goreleaser build --single-target --rm-dist --snapshot

test:
	@echo "==> Starting unit tests"
	go test $(TEST) -timeout=30s -parallel=4

attach:
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient attach $$(pgrep terraform-provider-artifactory)

acceptance: fmtcheck
	export TF_ACC=true && \
		go test -ldflags="-X '${PKG_VERSION_PATH}.Version=${NEXT_VERSION}-test'" -v -p 1 -parallel 20 -timeout 20m ./pkg/...

acceptance_federated:
	export TF_ACC=true && \
		go test -v -run TestAccFederatedRepo ./pkg/...

fmt:
	@echo "==> Fixing source code with gofmt..."
	@gofmt -s -w ./$(PKG_NAME)
	(command -v ${GOBIN}/goimports &> /dev/null || go get golang.org/x/tools/cmd/goimports) && ${GOBIN}/goimports -w pkg/artifactory

fmtcheck:
	@echo "==> Checking that code complies with gofmt requirements..."
	@sh -c "find . -name '*.go' -not -name '*vendor*' -print0 | xargs -0 gofmt -l -s"

.PHONY: build fmt
