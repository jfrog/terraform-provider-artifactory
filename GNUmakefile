TEST?=./pkg/...
PRODUCT=artifactory
GO_ARCH=$(shell go env GOARCH)
TARGET_ARCH=$(shell go env GOOS)_${GO_ARCH}
GORELEASER_ARCH=${TARGET_ARCH}
LINUX_GORELEASER_ARCH=linux_${GO_ARCH}

ifeq ($(GO_ARCH), amd64)
GORELEASER_ARCH=${TARGET_ARCH}_$(shell go env GOAMD64)
LINUX_GORELEASER_ARCH:=${LINUX_GORELEASER_ARCH}_$(shell go env GOAMD64)
endif

PKG_NAME=pkg/artifactory
# if this path ever changes, you need to also update the 'ldflags' value in .goreleaser.yml
PROVIDER_VERSION?=$(shell git describe --tags --abbrev=0 | sed  -n 's/v\([0-9]*\).\([0-9]*\).\([0-9]*\)/\1.\2.\3/p')
PROVIDER_MAJOR_VERSION?=$(shell echo ${PROVIDER_VERSION}| awk -F '.' '{print $$1}' )
NEXT_PROVIDER_VERSION := $(shell echo ${PROVIDER_VERSION}| awk -F '.' '{print $$1 "." $$2 "." $$3 +1 }' )
PKG_VERSION_PATH=github.com/jfrog/terraform-provider-${PRODUCT}/v${PROVIDER_MAJOR_VERSION}/${PKG_NAME}

TERRAFORM_CLI?=terraform

REGISTRY_HOST=registry.terraform.io
TF_ACC_PROVIDER_NAMESPACE=hashicorp

ifeq ($(TERRAFORM_CLI), tofu)
REGISTRY_HOST=registry.opentofu.org
TF_ACC_TERRAFORM_PATH="$(which tofu)"
TF_ACC_PROVIDER_HOST="registry.opentofu.org"
endif

BUILD_PATH=terraform.d/plugins/${REGISTRY_HOST}/jfrog/${PRODUCT}/${NEXT_PROVIDER_VERSION}/${TARGET_ARCH}
LINUX_BUILD_PATH=terraform.d/plugins/${REGISTRY_HOST}/jfrog/${PRODUCT}/${NEXT_PROVIDER_VERSION}/linux_amd64

SONAR_SCANNER_VERSION?=4.7.0.2747
SONAR_SCANNER_HOME?=${HOME}/.sonar/sonar-scanner-${SONAR_SCANNER_VERSION}-macosx

SMOKE_TESTS=(TestAccDataSourceUser_basic|TestAccLocalGenericRepository|TestAccLocalGenericRepositoryWithProjectAttributes|TestAccRemoteRepository_basic|TestAccRemoteRepositoryWithProjectAttributes|TestAccVirtualRepository_basic|TestAccVirtualGenericRepositoryWithProjectAttributes|TestAccFederatedRepoWithMembers|TestAccFederatedRepoWithProjectAttributes|TestAccWebhookAllTypes|TestAccUser_basic|TestAccGroup_basic|TestAccScopedToken_WithDefaults|TestAccPermissionTarget_full|TestAccBackup_full|TestAccGeneralSecurity_full|TestAccLdapGroupSetting_full|TestAccLdapSetting_full|TestAccOauthSettings_full|TestAccPropertySet|TestAccProxy|TestAccLayout_full|TestAccSamlSettings_full)

default: build

install: clean build
	mkdir -p ${BUILD_PATH} && \
	mkdir -p ${LINUX_BUILD_PATH} && \
		mv -v dist/terraform-provider-${PRODUCT}_${GORELEASER_ARCH}/terraform-provider-${PRODUCT}_v${NEXT_PROVIDER_VERSION}* ${BUILD_PATH} && \
		sed -i.bak 's/version = ".*"/version = "${NEXT_PROVIDER_VERSION}"/' sample.tf && rm sample.tf.bak && \
		${TERRAFORM_CLI} init

		# move this line up when testing on TFC
		# mv -v dist/terraform-provider-${PRODUCT}_${LINUX_GORELEASER_ARCH}/terraform-provider-${PRODUCT}_v${NEXT_PROVIDER_VERSION}* ${LINUX_BUILD_PATH} && \

clean:
	rm -fR dist terraform.d/ .terraform terraform.tfstate* terraform.d/ .terraform.lock.hcl

release:
	@git tag ${NEXT_PROVIDER_VERSION} && git push --mirror
	@echo "Pushed ${NEXT_PROVIDER_VERSION}"
	GOPROXY=https://proxy.golang.org GO111MODULE=on go get github.com/jfrog/terraform-provider-${PRODUCT}@v${NEXT_PROVIDER_VERSION}
	@echo "Updated pkg cache"

update_pkg_cache:
	GOPROXY=https://proxy.golang.org GO111MODULE=on go get github.com/jfrog/terraform-provider-${PRODUCT}@v${PROVIDER_VERSION}

build: fmt
	GORELEASER_CURRENT_TAG=${NEXT_PROVIDER_VERSION} goreleaser build --clean --snapshot --single-target

test:
	@echo "==> Starting unit tests"
	go test $(TEST) -timeout=30s -parallel=4

attach:
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient attach $$(pgrep terraform-provider-${PRODUCT})

smoke: fmt
	export TF_ACC=true && \
		go test -run '${SMOKE_TESTS}' -ldflags="-X '${PKG_VERSION_PATH}/provider.Version=${NEXT_PROVIDER_VERSION}-test'" -v -p 1 -timeout 5m $(TEST). -count=1

acceptance: fmt
	export TF_ACC=true && \
		go test -cover -coverprofile=coverage.txt -ldflags="-X '${PKG_VERSION_PATH}/provider.Version=${NEXT_PROVIDER_VERSION}-test'" -v -p 1 -parallel 20 -timeout 1h $(TEST)

coverage:
	go tool cover -html=coverage.txt

scan:
	${SONAR_SCANNER_HOME}/bin/sonar-scanner -Dsonar.projectVersion=${PROVIDER_VERSION} -Dsonar.go.coverage.reportPaths=coverage.txt

fmt:
	@echo "==> Fixing source code with gofmt..."
	@go fmt ./pkg/...

doc:
	go generate

.PHONY: build fmt
