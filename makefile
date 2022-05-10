MAIN_BRANCH := main
HEAD_BRANCH := HEAD
ifeq ($(strip $(VERSION_HASH)),)
# hash of current commit
VERSION_HASH := $(shell git rev-parse --short HEAD)
# tag matching current commit or empty
HEAD_TAG := $(shell git tag --points-at HEAD)
#name of branch
BRANCH_NAME := $(shell git rev-parse --abbrev-ref HEAD)
endif

VERSION_STRING := $(BRANCH_NAME)
#if we are on main and there is a tag pointing at head, use that for version else use branch name as version
ifeq ($(BRANCH_NAME),$(MAIN_BRANCH))
$(info "match main")
ifneq ($(strip $(HEAD_TAG)),)
VERSION_STRING := $(HEAD_TAG)
$(info    $(VERSION_STRING))
endif
endif

#if we are on HEAD and there is a tag pointing at head, use that for version else use branch name as version
ifeq ($(BRANCH_NAME),$(HEAD_BRANCH))
$(info match head)
ifneq ($(strip $(HEAD_TAG)),)
VERSION_STRING := $(HEAD_TAG)
$(info    $(version_string))
endif
endif


BINDIR    := $(CURDIR)/bin
PLATFORMS := linux/amd64/rk-Linux-x86_64 darwin/amd64/rk-Darwin-x86_64 windows/amd64/rk.exe linux/arm64/rk-Linux-arm64 darwin/arm64/rk-Darwin-arm64
# dlv exec ./bin/previewd --headless --listen=:2345 --log --api-version=2 -- testserver --loglevel=debug
BUILDCOMMANDDEBUG := go build -gcflags "all=-N -l" -tags "osusergo netgo static_build"
BUILDCOMMAND := go build -trimpath -ldflags "-s -w -X github.com/clarkezone/previewd/pkg/config.VersionHash=${VERSION_HASH} -X github.com/clarkezone/previewd/pkg/config.VersionString=${VERSION_STRING}" -tags "osusergo netgo static_build"
temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
label = $(word 3, $(temp))

UNAME := $(shell uname)
ifeq ($(UNAME), Darwin)
SHACOMMAND := shasum -a 256
else
SHACOMMAND := sha256sum
endif

.DEFAULT_GOAL := build

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
	go install github.com/uw-labs/strongbox@latest

.PHONY: test
test:
	./scripts/setenv.sh
	go test -p 4 -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: integration-actions
integration-actions:
	go test g -tags="common actions" --count=1 -v -timeout 15m

.PHONY: dep
dep:
	go mod tidy

.PHONY: latest
latest:
	echo ${VERSION_STRING} > bin/latest

.PHONY: lint
lint:
	revive $(shell go list ./...)
	go vet $(shell go list ./...)
	golangci-lint run

.PHONY: precommit
precommit:
	pre-commit run --all-files

.PHONY: build
build:
	$(BUILDCOMMAND) -o ${BINDIR}/previewd

.PHONY: builddlv
builddlv:
	$(BUILDCOMMANDDEBUG) -o ${BINDIR}/previewd

.PHONY: release
build-all: $(PLATFORMS)

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) CGO_ENABLED=0 $(BUILDCOMMAND) -o "bin/$(label)"
	$(SHACOMMAND) "bin/$(label)" > "bin/$(label).sha256"
