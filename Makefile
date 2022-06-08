SHELL = /bin/bash
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOPATH ?= $(shell go env GOPATH)
ROOT_DIRECTORY := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
MODULES_DIRECTORY := $(ROOT_DIRECTORY)/modules
GATEWAY_DIRECTORY := $(ROOT_DIRECTORY)/hack/gateway
COVERAGE_DIRECTORY = $(ROOT_DIRECTORY)/coverage
GO_COVERAGE_FILE = $(ROOT_DIRECTORY)/coverage/go.txt
AIR_BINARY := $(shell which air)
CODEGEN_VERSION := v0.24.0
CODEGEN_BIN := $(GOPATH)/pkg/mod/k8s.io/code-generator@$(CODEGEN_VERSION)/generate-groups.sh
MAIN_PACKAGE = github.com/kubernetes/dashboard/src/app/backend
KUBECONFIG ?= $(HOME)/.kube/config
SIDECAR_HOST ?= http://localhost:8000
TOKEN_TTL ?= 900
AUTO_GENERATE_CERTS ?= false
BIND_ADDRESS ?= 127.0.0.1
PORT ?= 8080
ENABLE_INSECURE_LOGIN ?= false
ENABLE_SKIP_LOGIN ?= false
SYSTEM_BANNER ?= "Local test environment"
SYSTEM_BANNER_SEVERITY ?= INFO
PROD_BINARY = .dist/amd64/web/dashboard
SERVE_DIRECTORY = .dist/web
SERVE_BINARY = .dist/web/dashboard
RELEASE_IMAGE = kubernetesui/dashboard
RELEASE_VERSION = v2.6.0
RELEASE_IMAGE_NAMES += $(foreach arch, $(ARCHITECTURES), $(RELEASE_IMAGE)-$(arch):$(RELEASE_VERSION))
RELEASE_IMAGE_NAMES_LATEST += $(foreach arch, $(ARCHITECTURES), $(RELEASE_IMAGE)-$(arch):latest)
HEAD_IMAGE = kubernetesdashboarddev/dashboard
HEAD_VERSION = latest
HEAD_IMAGE_NAMES += $(foreach arch, $(ARCHITECTURES), $(HEAD_IMAGE)-$(arch):$(HEAD_VERSION))
ARCHITECTURES = amd64 arm64 arm ppc64le s390x

# List of targets that should be always executed before other targets
PRE = --ensure-tools

.PHONY: check-license
check-license: $(PRE)
	@${GOPATH}/bin/license-eye header check

.PHONY: fix-license
fix-license: $(PRE)
	@${GOPATH}/bin/license-eye header fix

.PHONY: serve
serve: $(PRE)
	@$(MAKE) --no-print-directory -C $(MODULES_DIRECTORY) serve

.PHONY: serve-https
serve-https: $(PRE)
	@$(MAKE) --no-print-directory -C $(MODULES_DIRECTORY) serve-https

.PHONY: run
run: build-docker
	docker run --rm -p 8001:8001 --net dashboard --name dashboard-web -d dashboard-web --locale-config /public/locale_conf.json --auto-generate-certificates
	docker run --rm -p 9000:9000 -v $(KUBECONFIG):/config --net dashboard --name dashboard-api -d dashboard-api --kubeconfig config --auto-generate-certificates
	docker run --rm -p 443:443 --net dashboard --name gateway -d gateway

.PHONY: build-docker
build-docker: --ensure-docker-network build
	@docker build -f hack/gateway/Dockerfile -t gateway .
	@docker build -f modules/api/Dockerfile --build-arg BUILDPLATFORM=linux/amd64 -t dashboard-api .
	@docker build -f modules/web/Dockerfile --build-arg BUILDPLATFORM=linux/amd64 -t dashboard-web .

.PHONY: build
build:
	@$(MAKE) --no-print-directory -C $(MODULES_DIRECTORY) build

.PHONY: --ensure-tools
--ensure-tools:
	@$(MAKE) --no-print-directory -C $(MODULES_DIRECTORY)/tools install

.PHONY: --ensure-docker-network
--ensure-docker-network:
	@docker network create dashboard || true

.PHONY: --ensure-certificates
--ensure-certificates:
	@$(MAKE) --no-print-directory -C $(GATEWAY_DIRECTORY) generate-certificates

#.PHONY: ensure-codegen
#ensure-codegen: ensure-go
#	go get -d k8s.io/code-generator@$(CODEGEN_VERSION)
#	go mod tidy
#	chmod +x $(CODEGEN_BIN)
#
#.PHONY: clean
#clean:
#	rm -rf .tmp
#
#.PHONY: build-cross
#build-cross: clean ensure-go
#	./aio/scripts/build.sh -c
#
#.PHONY: prod-backend
#prod-backend: clean ensure-go
#	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X $(MAIN_PACKAGE)/client.Version=$(RELEASE_VERSION)" -o $(PROD_BINARY) $(MAIN_PACKAGE)
#
#.PHONY: prod-backend-cross
#prod-backend-cross: clean ensure-go
#	for ARCH in $(ARCHITECTURES) ; do \
#  	CGO_ENABLED=0 GOOS=linux GOARCH=$$ARCH go build -a -installsuffix cgo -ldflags "-X $(MAIN_PACKAGE)/client.Version=$(RELEASE_VERSION)" -o dist/$$ARCH/dashboard $(MAIN_PACKAGE) ; \
#  done
#
#.PHONY: prod
#prod: build
#	$(PROD_BINARY) --kubeconfig=$(KUBECONFIG) \
#		--sidecar-host=$(SIDECAR_HOST) \
#		--auto-generate-certificates \
#		--locale-config=dist/amd64/locale_conf.json \
#		--bind-address=${BIND_ADDRESS} \
#		--port=${PORT}
#
#.PHONY: test-backend
#test-backend: ensure-go
#	go test $(MAIN_PACKAGE)/...
#
#.PHONY: test-frontend
#test-frontend:
#	npx jest -c aio/jest.config.js
#
#.PHONY: test
#test: test-backend test-frontend
#
#.PHONY: coverage-backend
#coverage-backend: ensure-go
#	$(shell mkdir -p $(COVERAGE_DIRECTORY)) \
#	go test -coverprofile=$(GO_COVERAGE_FILE) -covermode=atomic $(MAIN_PACKAGE)/...
#
#.PHONY: coverage-frontend
#coverage-frontend:
#	npx jest -c aio/jest.config.js --coverage -i
#
#.PHONY: coverage
#coverage: coverage-backend coverage-frontend
#
#.PHONY: check-i18n
#check-i18n: fix-i18n
#
#.PHONY: fix-i18n
#fix-i18n:
#	./aio/scripts/pre-commit-i18n.sh
#
#.PHONY: check-codegen
#check-codegen: ensure-codegen
#	./aio/scripts/verify-codegen.sh
#
#.PHONY: fix-codegen
#fix-codegen: ensure-codegen
#	./aio/scripts/update-codegen.sh
#
#.PHONY: check-go
#check-go: ensure-golangcilint
#	golangci-lint run -c .golangci.yml ./src/app/backend/...
#
#.PHONY: fix-go
#fix-go: ensure-golangcilint
#	golangci-lint run -c .golangci.yml --fix ./src/app/backend/...
#
#.PHONY: check-html
#check-html:
#	./aio/scripts/check-html.sh
#
#.PHONY: fix-html
#fix-html:
#	npx html-beautify -f=./src/**/*.html
#
#.PHONY: check-scss
#check-scss:
#	stylelint "src/**/*.scss"
#
#.PHONY: fix-scss
#fix-scss:
#	stylelint "src/**/*.scss" --fix
#
#.PHONY: check-ts
#check-ts:
#	gts lint
#
#.PHONY: fix-ts
#fix-ts:
#	gts fix
#
#.PHONY: check-backend
#check-backend: check-license check-go check-codegen
#
#.PHONY: fix-backend
#fix-backend: fix-license fix-go fix-codegen
#
#.PHONY: check-frontend
#check-frontend: check-i18n check-license check-html check-scss check-ts
#
#.PHONY: fix-frontend
#fix-frontend: fix-i18n fix-license fix-html fix-scss fix-ts
#
#.PHONY: check
#check: check-i18n check-license check-go check-codegen check-html check-scss check-ts
#
#.PHONY: fix
#fix: fix-i18n fix-license fix-go fix-codegen  fix-html fix-scss fix-ts
#
#.PHONY: start-cluster
#start-cluster:
#	./aio/scripts/start-cluster.sh
#
#.PHONY: stop-cluster
#stop-cluster:
#	./aio/scripts/stop-cluster.sh
#
#.PHONY: e2e
#e2e: start-cluster
#	npm run e2e
#	make stop-cluster
#
#.PHONY: e2e-headed
#e2e-headed: start-cluster
#	npm run e2e:headed
#	make stop-cluster
#
#.PHONY: docker-build-release
#docker-build-release: build-cross
#	for ARCH in $(ARCHITECTURES) ; do \
#		docker buildx build \
#			-t $(RELEASE_IMAGE)-$$ARCH:$(RELEASE_VERSION) \
#			-t $(RELEASE_IMAGE)-$$ARCH:latest \
#			--build-arg BUILDPLATFORM=linux/$$ARCH \
#			--platform linux/$$ARCH \
#			--push \
#			dist/$$ARCH ; \
#	done ; \
#
#.PHONY: docker-push-release
#docker-push-release: docker-build-release
#	docker manifest create --amend $(RELEASE_IMAGE):$(RELEASE_VERSION) $(RELEASE_IMAGE_NAMES) ; \
#  docker manifest create --amend $(RELEASE_IMAGE):latest $(RELEASE_IMAGE_NAMES_LATEST) ; \
#  docker manifest push $(RELEASE_IMAGE):$(RELEASE_VERSION) ; \
#  docker manifest push $(RELEASE_IMAGE):latest
#
#.PHONY: docker-build-head
#docker-build-head: build-cross
#	for ARCH in $(ARCHITECTURES) ; do \
#		docker buildx build \
#			-t $(HEAD_IMAGE)-$$ARCH:$(HEAD_VERSION) \
#			--build-arg BUILDPLATFORM=linux/$$ARCH \
#			--platform linux/$$ARCH \
#			--push \
#			dist/$$ARCH ; \
#	done ; \
#
#.PHONY: docker-push-head
#docker-push-head: docker-build-head
#	docker manifest create --amend $(HEAD_IMAGE):$(HEAD_VERSION) $(HEAD_IMAGE_NAMES)
#	docker manifest push $(HEAD_IMAGE):$(HEAD_VERSION) ; \
