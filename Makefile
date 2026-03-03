APP     := moralink-gost
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.1")

# Load .env file if it exists — exports all vars into the make environment
ifneq (,$(wildcard .env))
  include .env
  export
endif

LDFLAGS := -ldflags "\
  -X main.Version=$(VERSION) \
  -X main.ReleaseGH=$(RELEASE_GH) \
  -s -w"

.PHONY: all linux windows clean tidy install-scheduler uninstall-scheduler

all: linux windows

## Build for Linux (amd64)
linux: dist
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
	go build $(LDFLAGS) -o dist/$(APP)-linux-amd64 .
	@echo "✓  Built dist/$(APP)-linux-amd64"

## Build for Windows (amd64)
## Requires mingw-w64 cross-compiler: apt install gcc-mingw-w64-x86-64
windows: dist
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
	go build $(LDFLAGS) -o dist/$(APP)-windows-amd64.exe .
	@echo "✓  Built dist/$(APP)-windows-amd64.exe"

## Register the updater in cron (Linux) or Task Scheduler (Windows)
install-scheduler:
	./dist/$(APP)-linux-amd64 --install-scheduler

## Remove the scheduled task
uninstall-scheduler:
	./dist/$(APP)-linux-amd64 --uninstall-scheduler

## Tidy and vendor
tidy:
	go mod tidy

## Create dist dir if missing
dist:
	@mkdir -p dist

clean:
	rm -rf dist/