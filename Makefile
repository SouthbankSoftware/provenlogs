# @Author: guiguan
# @Date:   2019-05-15T14:42:41+10:00
# @Last modified by:   guiguan
# @Last modified time: 2019-05-15T14:47:53+10:00

PROJECT_IMPORT_PATH := github.com/SouthbankSoftware/provenlogs
APP_NAME := provenlogs
APP_VERSION ?= 0.0.0
PLAYGROUND_NAME := playground
PKGS := $(shell go list ./cmd/... ./pkg/...)
LD_FLAGS := -ldflags \
"-X main.cmdVersion=$(APP_VERSION)"

all: build

.PHONY: run build test test-dev clean playground doc build-all

run:
	go run $(LD_FLAGS) ./cmd/$(APP_NAME) -h
build:
	go build $(LD_FLAGS) ./cmd/$(APP_NAME)
test:
	go test $(LD_FLAGS) $(PKGS)
test-dev:
	# -test.v verbose
	go test $(LD_FLAGS) -count=1 -test.v $(PKGS)
clean:
	go clean -testcache $(PKGS)
	rm -f $(APP_NAME)* $(PLAYGROUND_NAME)*
playground:
	go run cmd/$(PLAYGROUND_NAME)/$(PLAYGROUND_NAME).go
doc:
	godoc -http=:6060
build-all:
	go run github.com/mitchellh/gox -osarch="linux/amd64 windows/amd64 darwin/amd64" $(LD_FLAGS) ./cmd/$(APP_NAME)
