SHELL = /bin/bash
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOPATH ?= $(shell go env GOPATH)
ROOT_DIRECTORY := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
COVERAGE_DIRECTORY = $(ROOT_DIRECTORY)/coverage
GO_COVERAGE_FILE = $(ROOT_DIRECTORY)/coverage/go.txt
FSWATCH_BINARY := $(shell which fswatch)
GOLANGCILINT_BINARY := $(shell which golangci-lint)
GOLANGCILINT_VERSION := v1.42.1
GO_BINARY := $(shell which go)
GO_MAJOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
MIN_GO_MAJOR_VERSION = 1
MIN_GO_MINOR_VERSION = 17
MAIN_PACKAGE = github.com/kubernetes/dashboard/src/app/backend
KUBECONFIG ?= $(HOME)/.kube/config
SIDECAR_HOST ?= http://localhost:8000
TOKEN_TTL ?= 900
AUTO_GENERATE_CERTS ?= false
ENABLE_INSECURE_LOGIN ?= false
ENABLE_SKIP_LOGIN ?= false
SYSTEM_BANNER ?= "Local test environment"
SYSTEM_BANNER_SEVERITY ?= INFO
PROD_BINARY = dist/amd64/dashboard
SERVE_DIRECTORY = .tmp/serve
SERVE_BINARY = .tmp/serve/dashboard
SERVE_PID = .tmp/serve/dashboard.pid
RELEASE_IMAGE = kubernetesui/dashboard
RELEASE_VERSION = v2.3.1
RELEASE_IMAGE_NAMES += $(foreach arch, $(ARCHITECTURES), $(RELEASE_IMAGE)-$(arch):$(RELEASE_VERSION))
RELEASE_IMAGE_NAMES_LATEST += $(foreach arch, $(ARCHITECTURES), $(RELEASE_IMAGE)-$(arch):latest)
HEAD_IMAGE = kubernetesdashboarddev/dashboard
HEAD_VERSION = head
HEAD_IMAGE_NAMES += $(foreach arch, $(ARCHITECTURES), $(HEAD_IMAGE)-$(arch):$(HEAD_VERSION))
ARCHITECTURES = amd64 arm64 arm ppc64le s390x

.PHONY: post-install
post-install: ensure-version
	go mod vendor
	./aio/scripts/install-codegen.sh
	npx ngcc

.PHONY: ensure-version
ensure-version:
	node ./aio/scripts/version.mjs

.PHONY: ensure-golangcilint
ensure-golangcilint:
ifndef GOLANGCILINT_BINARY
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin $(GOLANGCILINT_VERSION)
endif

.PHONY: ensure-fswatch
ensure-fswatch:
ifndef FSWATCH_BINARY
	$(error "Cannot find fswatch binary")
endif

.PHONY: ensure-go
ensure-go:
ifndef GO_BINARY
	$(error "Cannot find go binary")
endif
	@if [ $(GO_MAJOR_VERSION) -gt $(MIN_GO_MAJOR_VERSION) ]; then \
		exit 0 ;\
  elif [ $(GO_MAJOR_VERSION) -lt $(MIN_GO_MAJOR_VERSION) ]; then \
		exit 1; \
  elif [ $(GO_MINOR_VERSION) -lt $(MIN_GO_MINOR_VERSION) ] ; then \
		exit 1; \
  fi

.PHONY: clean
clean:
	rm -rf .tmp

.PHONY: build-backend
build-backend: ensure-go
	go build -ldflags "-X $(MAIN_PACKAGE)/client.Version=$(RELEASE_VERSION)" -gcflags="all=-N -l" -o $(SERVE_BINARY) $(MAIN_PACKAGE)

.PHONY: build
build: clean ensure-go
	./aio/scripts/build.sh

.PHONY: build-cross
build-cross: clean ensure-go
	./aio/scripts/build.sh -c

.PHONY: serve-backend
serve-backend: build-backend
	cp i18n/locale_conf.json $(SERVE_DIRECTORY)/locale_conf.json
	$(SERVE_BINARY) --kubeconfig=$(KUBECONFIG) \
		--sidecar-host=$(SIDECAR_HOST) \
		--system-banner=$(SYSTEM_BANNER) \
		--system-banner-severity=$(SYSTEM_BANNER_SEVERITY) \
		--token-ttl=$(TOKEN_TTL) \
		--auto-generate-certificates=$(AUTO_GENERATE_CERTS) \
		--enable-insecure-login=$(ENABLE_INSECURE_LOGIN) \
		--enable-skip-login=$(ENABLE_SKIP_LOGIN) & echo $$! > $(SERVE_PID)

.PHONY: kill-backend
kill-backend:
	kill `cat $(SERVE_PID)` || true
	rm -rf $(SERVE_PID)

.PHONY: restart-backend
restart-backend: kill-backend serve-backend

.PHONY: watch-backend
watch-backend: ensure-fswatch restart-backend
	fswatch -o -r -e '.*' -i '\.go$$'  . | xargs -n1 -I{} make restart-backend || make kill-backend

.PHONY: prod-backend
prod-backend: clean ensure-go
	GOOS=linux go build -a -installsuffix cgo -ldflags "-X $(MAIN_PACKAGE)/client.Version=$(RELEASE_VERSION)" -o $(PROD_BINARY) $(MAIN_PACKAGE)

.PHONY: prod-backend-cross
prod-backend-cross: clean ensure-go
	for ARCH in $(ARCHITECTURES) ; do \
  	GOOS=linux GOARCH=$$ARCH go build -a -installsuffix cgo -ldflags "-X $(MAIN_PACKAGE)/client.Version=$(RELEASE_VERSION)" -o dist/$$ARCH/dashboard $(MAIN_PACKAGE) ; \
  done

.PHONY: prod
prod: build
	$(PROD_BINARY) --kubeconfig=$(KUBECONFIG) \
		--sidecar-host=$(SIDECAR_HOST) \
		--auto-generate-certificates \
		--locale-config=dist/amd64/locale_conf.json \
		--bind-address=127.0.0.1 \
		--port=8080

.PHONY: test-backend
test-backend: ensure-go
	go test $(MAIN_PACKAGE)/...

.PHONY: test-frontend
test-frontend:
	npm run test

.PHONY: test
test: test-backend test-frontend

.PHONY: coverage-backend
coverage-backend: ensure-go
	$(shell mkdir -p $(COVERAGE_DIRECTORY)) \
	go test -coverprofile=$(GO_COVERAGE_FILE) -covermode=atomic $(MAIN_PACKAGE)/...

.PHONY: coverage-frontend
coverage-frontend:
	npm run test:coverage

.PHONY: coverage
coverage: coverage-backend coverage-frontend

.PHONY: check-i18n
check-i18n: fix-i18n

.PHONY: fix-i18n
fix-i18n:
	./aio/scripts/pre-commit-i18n.sh

.PHONY: check-license
check-license:
	license-check-and-add check -f license-checker-config.json

.PHONY: fix-license
fix-license:
	license-check-and-add add -f license-checker-config.json

.PHONY: check-codegen
check-codegen:
	./aio/scripts/verify-codegen.sh

.PHONY: fix-codegen
fix-codegen:
	./aio/scripts/update-codegen.sh

.PHONY: check-go
check-go: ensure-golangcilint
	golangci-lint run -c .golangci.yml ./src/app/backend/...

.PHONY: fix-go
fix-go: ensure-golangcilint
	golangci-lint run -c .golangci.yml --fix ./src/app/backend/...

.PHONY: check-html
check-html:
	./aio/scripts/format-html.sh --check

.PHONY: fix-html
fix-html:
	./aio/scripts/format-html.sh

.PHONY: check-scss
check-scss:
	stylelint "src/**/*.scss"

.PHONY: fix-scss
fix-scss:
	stylelint "src/**/*.scss" --fix

.PHONY: check-ts
check-ts:
	gts lint

.PHONY: fix-ts
fix-ts:
	gts fix

.PHONY: check
check: check-i18n check-license check-codegen check-go check-html check-scss check-ts

.PHONY: fix
check: fix-i18n fix-license fix-codegen fix-go fix-html fix-scss fix-ts

.PHONY: start-cluster
start-cluster:
	./aio/scripts/start-cluster.sh

.PHONY: stop-cluster
stop-cluster:
	./aio/scripts/stop-cluster.sh

.PHONY: e2e
e2e: start-cluster
	npm run e2e
	make stop-cluster

.PHONY: docker-build-release
docker-build-release: build-cross
	for ARCH in $(ARCHITECTURES) ; do \
  		docker build -t $(RELEASE_IMAGE)-$$ARCH:$(RELEASE_VERSION) -t $(RELEASE_IMAGE)-$$ARCH:latest dist/$$ARCH ; \
  done

.PHONY: docker-push-release
docker-push-release: docker-build-release
	for ARCH in $(ARCHITECTURES) ; do \
  		docker push $(RELEASE_IMAGE)-$$ARCH:$(RELEASE_VERSION) ; \
  		docker push $(RELEASE_IMAGE)-$$ARCH:latest ; \
  done ; \
  docker manifest create --amend $(RELEASE_IMAGE):$(RELEASE_VERSION) $(RELEASE_IMAGE_NAMES) ; \
  docker manifest create --amend $(RELEASE_IMAGE):latest $(RELEASE_IMAGE_NAMES_LATEST) ; \
	for ARCH in $(ARCHITECTURES) ; do \
  		docker manifest annotate $(RELEASE_IMAGE):$(RELEASE_VERSION) $(RELEASE_IMAGE)-$$ARCH:$(RELEASE_VERSION) --os linux --arch $$ARCH ; \
  		docker manifest annotate $(RELEASE_IMAGE):latest $(RELEASE_IMAGE)-$$ARCH:latest --os linux --arch $$ARCH ; \
  done ; \
  docker manifest push $(RELEASE_IMAGE):$(RELEASE_VERSION) ; \
  docker manifest push $(RELEASE_IMAGE):latest

.PHONY: docker-build-head
docker-build-head: build-cross
	for ARCH in $(ARCHITECTURES) ; do \
  		docker build -t $(HEAD_IMAGE)-$$ARCH:$(HEAD_VERSION) dist/$$ARCH ; \
  done

.PHONY: docker-push-head
docker-push-head: docker-build-head
	for ARCH in $(ARCHITECTURES) ; do \
  		docker push $(HEAD_IMAGE)-$$ARCH:$(HEAD_VERSION) ; \
  done ; \
  docker manifest create --amend $(HEAD_IMAGE):$(HEAD_VERSION) $(HEAD_IMAGE_NAMES)
	for ARCH in $(ARCHITECTURES) ; do \
  		docker manifest annotate $(HEAD_IMAGE):$(HEAD_VERSION) $(HEAD_IMAGE)-$$ARCH:$(HEAD_VERSION) --os linux --arch $$ARCH ; \
  done ; \
  docker manifest push $(HEAD_IMAGE):$(HEAD_VERSION)
