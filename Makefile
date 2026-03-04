APP     := moralink-gost
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.1")

ifneq (,$(wildcard .env))
  include .env
  export
endif

LDFLAGS := -ldflags "\
  -X main.Version=$(VERSION) \
  -X main.ReleaseGH=$(RELEASE_GH) \
  -s -w"

.PHONY: all linux windows clean tidy syso

all: linux windows

linux: dist
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
	go build $(LDFLAGS) -o dist/$(APP)-linux-amd64 .
	@echo "✓  Built dist/$(APP)-linux-amd64"

## Generates resource.syso — Go picks this up automatically when building for Windows
GOVERSIONINFO := $(shell go env GOPATH)/bin/goversioninfo

syso: $(GOVERSIONINFO)
	$(GOVERSIONINFO) -o resource.syso -platform-specific=true versioninfo.json
	@echo "✓  Generated resource.syso"

$(GOVERSIONINFO):
	@echo "→  Installing goversioninfo..."
	go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest

windows: dist syso
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
	go build $(LDFLAGS) -o dist/$(APP)-windows-amd64.exe .
	@echo "✓  Built dist/$(APP)-windows-amd64.exe"

install-scheduler:
	./dist/$(APP)-linux-amd64 --install-scheduler

uninstall-scheduler:
	./dist/$(APP)-linux-amd64 --uninstall-scheduler

tidy:
	go mod tidy

dist:
	@mkdir -p dist

clean:
	rm -rf dist/ resource.syso resource_windows_amd64.syso