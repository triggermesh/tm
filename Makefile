PACKAGE       = tm
PACKAGE_DESC  = Triggermesh CLI
TARGETS      ?= darwin/amd64 linux/amd64 windows/amd64
DIST_DIR     ?=

GIT_REPO      = github.com/triggermesh/tm
GIT_TAG      ?= $(shell git describe --tags --always)

DOCKER        = docker
IMAGE_REPO   ?= gcr.io/triggermesh
IMAGE        ?= $(IMAGE_REPO)/$(shell basename $(GIT_REPO)):$(GIT_TAG)

GO           ?= go
GOFMT        ?= gofmt
GOLINT       ?= golint
GOTEST       ?= gotestsum --junitfile $(OUTPUT_DIR)$(PACKAGE)-unit-tests.xml --format pkgname-and-test-fails --
GOTOOL       ?= go tool
GLIDE        ?= glide

LDFLAGS      += -s -w -X $(GIT_REPO)/cmd.version=${GIT_TAG}

.PHONY: help mod-download build release install test coverage lint vet fmt fmt-test image clean

all: build

help: ## Display this help
	@awk 'BEGIN {FS = ":.*?## "; printf "\n$(PACKAGE_DESC)\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9._-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

mod-download: ## Download go modules
	$(GO) mod download

build: ## Build the tm binary
	$(GO) build -v -mod=readonly -ldflags="$(LDFLAGS)" $(GIT_REPO)

install: ## Install the binary
	$(GO) install -v -mod=readonly -ldflags="$(LDFLAGS)" $(GIT_REPO)

release: ## Build release binaries
	@set -e ; \
	for platform in $(TARGETS); do \
		GOOS=$${platform%/*} ; \
		GOARCH=$${platform#*/} ; \
		RELEASE_BINARY=$(PACKAGE)-$${GOOS}-$${GOARCH} ; \
		[ $${GOOS} = "windows" ] && RELEASE_BINARY=$${RELEASE_BINARY}.exe ; \
		echo "GOOS=$${GOOS} GOARCH=$${GOARCH} $(GO) build -v -mod=readonly -ldflags="$(LDFLAGS)" -o $(DIST_DIR)$${RELEASE_BINARY} $(GIT_REPO)" ; \
		GOOS=$${GOOS} GOARCH=$${GOARCH} $(GO) build -v -mod=readonly -ldflags="$(LDFLAGS)" -o $(DIST_DIR)$${RELEASE_BINARY} $(GIT_REPO) ; \
	done

test: ## Run unit tests
	$(GOTEST) -v -timeout 15m -p=1 -cover -coverprofile=c.out $(GIT_REPO)/...

coverage: ## Generate code coverage
	$(GOTOOL) cover -html=c.out -o $(OUTPUT_DIR)$(PACKAGE)-coverage.html

lint: ## Link source files
	$(GOLINT) $(shell $(GLIDE) novendor)

vet: ## Vet source files
	$(GO) vet $(GIT_REPO)/...

fmt: ## Format source files
	$(GOFMT) -s -w $(shell $(GO) list -f '{{$$d := .Dir}}{{range .GoFiles}}{{$$d}}/{{.}} {{end}} {{$$d := .Dir}}{{range .TestGoFiles}}{{$$d}}/{{.}} {{end}}' $(GIT_REPO)/...)

fmt-test: ## Check source formatting
	@test -z $(shell $(GOFMT) -l $(shell $(GO) list -f '{{$$d := .Dir}}{{range .GoFiles}}{{$$d}}/{{.}} {{end}} {{$$d := .Dir}}{{range .TestGoFiles}}{{$$d}}/{{.}} {{end}}' $(GIT_REPO)/...))

image: ## Builds the container image
	$(DOCKER) build . -t $(IMAGE) --build-arg GIT_TAG=$(GIT_TAG)

clean: ## Clean build artifacts
	$(RM) -rf $(PACKAGE)
	$(RM) -rf $(PACKAGE)-unit-tests.xml
	$(RM) -rf c.out $(PACKAGE)-coverage.html
	@for platform in $(TARGETS); do $(RM) -rf $(DIST_DIR)$(PACKAGE)-$${platform%/*}-$${platform#*/}; done
