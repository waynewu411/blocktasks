GO_VERSION := $(shell grep '^go ' go.mod | awk '{print $$2}')
ARCH ?= $(if $(filter $(shell uname -m),x86_64),amd64,arm64)
VERSION ?= $(shell git rev-parse --short HEAD)
BUILD ?= $(shell date +%Y%m%d%H%M%S)
ALL_SRC_FILES := $(shell find . -type f -name '*.go' | sort)
GO := go
DOCKER := docker
MAKE := make
GOFMT := gofmt
GOIMPORTS := goimports
GCI := gci
RUNNER ?= gcr.io/distroless/base-debian12:nonroot
REGISTRY ?= docker.io/waynewu411
TARGET ?= blocktasks

.PHONY: all-tools
all-tools:
	@for tool in $(shell cat tools.go | grep _ | awk -F'"' '{print $$2}' | sed 's/^_ //'); do \
		echo "Installing $$tool"; \
		$(GO) install $$tool; \
	done

.PHONY: all-fmt
all-fmt:
	@$(GOFMT) -w -s ./
	@$(GOIMPORTS) -w $(ALL_SRC_FILES)
	@$(GCI) write -s standard -s default -s dot ./

.PHONY: all-tidy
all-tidy:
	@$(GO) mod tidy

.PHONY: all-mod-download
all-mod-download:
	@$(GO) mod download

.PHONY: build-target
build-target:
	$(GO) build -ldflags "-X 'main.Version=$(VERSION)' -X main.Build=$(BUILD)" -o cmd/app cmd/main.go

.PHONY: build-blocktasks
build-blocktasks:
	$(MAKE) build-target TARGET=blocktasks

.PHONY: docker-build-target
docker-build-target:
	DOCKER_BUILDKIT=1 $(DOCKER) build \
		--platform linux/$(ARCH) \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD=$(BUILD) \
		--build-arg RUNNER=$(RUNNER) \
		-f cmd/Dockerfile \
		-t $(REGISTRY)/$(TARGET):$(ARCH)-$(VERSION) \
		-t $(REGISTRY)/$(TARGET):$(ARCH)-latest \
		.

.PHONY: docker-build-blocktasks
docker-build-blocktasks:
	@$(MAKE) docker-build-target TARGET="blocktasks" RUNNER="gcr.io/distroless/base-debian12:nonroot"

.PHONY: docker-build-blocktasks-amd64
docker-build-blocktasks-amd64:
	@$(MAKE) docker-build-blocktasks ARCH="amd64" RUNNER="gcr.io/distroless/base-debian12:nonroot"

.PHONY: docker-build-blocktasks-arm64
docker-build-blocktasks-arm64:
	@$(MAKE) docker-build-blocktasks ARCH="arm64" RUNNER="gcr.io/distroless/base-debian12:nonroot"

.PHONY: docker-publish-image
docker-publish-image:
	docker push $(REGISTRY)/$(TARGET):$(ARCH)-$(VERSION)
	docker push $(REGISTRY)/$(TARGET):$(ARCH)-latest

.PHONY: docker-publish-blocktasks
docker-publish-blocktasks:
	@$(MAKE) docker-publish-image TARGET="blocktasks"

.PHONY: docker-publish-blocktasks-amd64
docker-publish-blocktasks-amd64:
	@$(MAKE) docker-publish-blocktasks ARCH="amd64"

.PHONY: docker-publish-blocktasks-arm64
docker-publish-blocktasks-arm64:
	@$(MAKE) docker-publish-blocktasks ARCH="arm64"
