SHELL := /bin/bash
OS := $(shell uname | tr '[:upper:]' '[:lower:]')

GO_VARS := GO111MODULE=on GO15VENDOREXPERIMENT=1 CGO_ENABLED=0
REV := $(shell git rev-parse --short HEAD 2> /dev/null || echo 'unknown')
IMAGE ?= gcr.io/jenkinsxio/jx-app-sonar-scanner
VERSION ?= 0.0.0-dev-$(REV)
BUILDFLAGS := '-X github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/version.binaryVersion=$(VERSION) -X github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/version.imageName=$(IMAGE)'

APP_NAME := jx-app-sonar-scanner
MAIN := cmd/sonar-scanner/main.go

BUILD_DIR=build
PACKAGE_DIRS := $(shell go list ./...)
PKGS := $(subst  :,_,$(PACKAGE_DIRS))
PLATFORMS := windows linux darwin
os = $(word 1, $@)

DOCKER_REGISTRY ?= docker.io

GOMMIT_START_SHA ?= 01b8d360a549e0a80b9fc9c587b69bba616e8d85

FGT := $(GOPATH)/bin/fgt
GOLINT := $(GOPATH)/bin/golint
GOMMIT := $(GOPATH)/bin/gommit

.PHONY : all
all: linux test check ## Compiles, test and verifies source

.PHONY: $(PLATFORMS)
$(PLATFORMS):	
	$(GO_VARS) GOOS=$(os) GOARCH=amd64 go build -ldflags $(BUILDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)

.PHONY : test
test: ## Runs unit tests
	$(GO_VARS) go test -coverprofile=coverage.out -v ./...

.PHONY : fmt
fmt: ## Re-formates Go source files according to standard
	@$(GO_VARS) go fmt ./...

.PHONY : clean
clean: ## Deletes the build directory with all generated artefacts
	rm -rf $(BUILD_DIR)

check: $(GOLINT) $(FGT) $(GOMMIT)
	@echo "LINTING"
	@$(FGT) $(GOLINT) ./...
	@echo "VETTING"
	@$(GO_VARS) $(FGT) go vet ./...
	#@echo "CONVENTIONAL COMMIT CHECK"
	#@$(GOMMIT) check range $(GOMMIT_START_SHA) $$(git log --pretty=format:'%H' -n 1)

.PHONY: watch
watch: ## Watches for file changes in Go source files and re-runs 'skaffold build'. Requires entr
	find . -name "*.go" | entr -s 'make skaffold-build' 

.PHONY: skaffold-build
skaffold-build: linux ## Runs 'skaffold build'
	DOCKER_REGISTRY=$(DOCKER_REGISTRY) VERSION=$(VERSION) skaffold build -f skaffold.yaml

.PHONY: skaffold-run
skaffold-run: linux ## Runs 'skaffold run'
	DOCKER_REGISTRY=$(DOCKER_REGISTRY) VERSION=$(VERSION) skaffold run -f skaffold.yaml -p dev

.PHONY: help
help: ## Prints this help
	@grep -E '^[^.]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'	

.PHONY: release
release: linux test check update-release-version ## skaffold-build detach-and-release ## Creates a release
	cd charts/jx-app-sonar-scanner && jx step helm release

.PHONY: update-release-version
update-release-version: ## Updates the release version
ifeq ($(OS),darwin)
	sed -i "" -e "s/version:.*/version: $(VERSION)/" ./charts/jx-app-sonar-scanner/Chart.yaml
	sed -i "" -e "s/tag: .*/tag: $(VERSION)/" ./charts/jx-app-sonar-scanner/values.yaml
else ifeq ($(OS),linux)
	sed -i -e "s/version:.*/version: $(VERSION)/" ./charts/jx-app-sonar-scanner/Chart.yaml
	sed -i -e "s/tag: .*/tag: $(VERSION)/" ./charts/jx-app-sonar-scanner/values.yaml
else
	echo "platform $(OS) not supported to tag with"
	exit -1
endif

.PHONY: detach-and-release
detach-and-release:  ## Gets into detached HEAD mode and  pushes release
	git checkout $(shell git rev-parse HEAD)
	git add --all
	git commit -m "release $(VERSION)" --allow-empty # if first release then no version update is performed
	git tag -fa v$(VERSION) -m "Release version $(VERSION)"
	git push origin HEAD v$(VERSION)

# Targets to get some Go tools
$(FGT):
	@$(GO_VARS) go get github.com/GeertJohan/fgt

$(GOLINT):
	@$(GO_VARS) go get golang.org/x/lint/golint

$(GOMMIT):
	@$(GO_VARS) go get github.com/antham/gommit
