SHELL=/bin/bash
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
ROOT_DIRECTORY := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
COVERAGE_DIRECTORY=$(ROOT_DIRECTORY)/coverage
GO_COVERAGE_FILE=$(ROOT_DIRECTORY)/coverage/go.txt
GO_BINARY := $(shell which go)
GO_MAJOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
MIN_GO_MAJOR_VERSION=1
MIN_GO_MINOR_VERSION=16
RELEASE_IMAGE=kubernetesui/dashboard
RELEASE_VERSION=v2.3.1
RELEASE_IMAGE_NAMES += $(foreach arch, $(ARCHITECTURES), $(RELEASE_IMAGE)-$(arch):$(RELEASE_VERSION))
RELEASE_IMAGE_NAMES_LATEST += $(foreach arch, $(ARCHITECTURES), $(RELEASE_IMAGE)-$(arch):latest)
HEAD_IMAGE=kubernetesdashboarddev/dashboard
HEAD_VERSION=head
HEAD_IMAGE_NAMES += $(foreach arch, $(ARCHITECTURES), $(HEAD_IMAGE)-$(arch):$(HEAD_VERSION))
ARCHITECTURES=amd64 arm64 arm ppc64le s390x

.PHONY: validate-go
validate-go:
ifndef GO_BINARY
	exit 1
endif
	@if [ $(GO_MAJOR_VERSION) -gt $(MIN_GO_MAJOR_VERSION) ]; then \
		exit 0 ;\
  elif [ $(GO_MAJOR_VERSION) -lt $(MIN_GO_MAJOR_VERSION) ]; then \
		exit 1; \
  elif [ $(GO_MINOR_VERSION) -lt $(MIN_GO_MINOR_VERSION) ] ; then \
		exit 1; \
  fi

.PHONY: validate
validate: validate-go

.PHONY: test-backend
test-backend: validate-go
	go test github.com/kubernetes/dashboard/src/app/backend/...

.PHONY: test-frontend
test-frontend:
	npm run test

.PHONY: test
test: test-backend test-frontend

.PHONY: coverage-backend
coverage-backend: validate-go
	$(shell mkdir -p $(COVERAGE_DIRECTORY)) \
	go test -coverprofile=$(GO_COVERAGE_FILE) -covermode=atomic github.com/kubernetes/dashboard/src/app/backend/...

.PHONY: coverage-frontend
coverage-frontend:
	npm run test:coverage

.PHONY: coverage
coverage: coverage-backend coverage-frontend

.PHONY: clean
clean:
	rm -rf .go_workspace .tmp coverage dist npm-debug.log

.PHONY: build-cross
build-cross: clean validate
	./aio/scripts/build.sh -c

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
