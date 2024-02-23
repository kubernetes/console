### Common application/container details
PROJECT_NAME := dashboard
# Supported architectures
ARCHITECTURES := linux/amd64 linux/arm64 linux/arm linux/ppc64le linux/s390x # darwin/amd64 darwin/arm64 <- TODO: enable once it is natively supported by docker
BUILDX_ARCHITECTURES := linux/amd64,linux/arm64,linux/arm,linux/ppc64le,linux/s390x # ,darwin/amd64,darwin/arm64
# Container registry details
IMAGE_REGISTRIES := docker.io # ghcr.io <- TODO: uncomment when ghcr will be supported
IMAGE_REPOSITORY := kubernetesui

### Dirs and paths
# Base paths
PARTIALS_DIRECTORY := $(ROOT_DIRECTORY)/hack/include
# Modules
MODULES_DIRECTORY := $(ROOT_DIRECTORY)/modules
API_DIRECTORY := $(MODULES_DIRECTORY)/api
AUTH_DIRECTORY := $(MODULES_DIRECTORY)/auth
METRICS_SCRAPER_DIRECTORY := $(MODULES_DIRECTORY)/metrics-scraper
WEB_DIRECTORY := $(MODULES_DIRECTORY)/web
TOOLS_DIRECTORY := $(MODULES_DIRECTORY)/common/tools
# Gateway
GATEWAY_DIRECTORY := $(ROOT_DIRECTORY)/hack/gateway
# Docker files
DOCKER_DIRECTORY := $(ROOT_DIRECTORY)/hack/docker
DOCKER_COMPOSE_PATH := $(DOCKER_DIRECTORY)/docker.compose.yaml
DOCKER_COMPOSE_DEV_PATH := $(DOCKER_DIRECTORY)/dev.compose.yml
# Build
DIST_DIRECTORY := $(ROOT_DIRECTORY)/.dist
TMP_DIRECTORY := $(ROOT_DIRECTORY)/.tmp
# Kind
KIND_CLUSTER_NAME := kubernetes-dashboard
KIND_CLUSTER_VERSION := 1.29.0
KIND_CLUSTER_IMAGE := docker.io/kindest/node:v${KIND_CLUSTER_VERSION}
KIND_CLUSTER_INTERNAL_KUBECONFIG_PATH := $(TMP_DIRECTORY)/kubeconfig

### GOPATH check
ifndef GOPATH
$(warning $$GOPATH environment variable not set)
endif

ifeq (,$(findstring $(GOPATH)/bin,$(PATH)))
$(warning $$GOPATH/bin directory is not in your $$PATH)
endif
