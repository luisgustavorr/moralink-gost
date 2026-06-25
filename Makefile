APP     := moralink-gost
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.1")

ifneq (,$(wildcard .env))
  include .env
  export
endif

LDFLAGS := -ldflags "\
  -X main.Version=$(VERSION) \
  -X main.ReleaseGH=$(RELEASE_GH) \
  -X main.SharkToken=$(SHARK_TOKEN) \
  -s -w"

.PHONY: all linux windows clean tidy syso

all: linux installer

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

windows32: dist syso
	GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-gcc \
	go build $(LDFLAGS) -o dist/$(APP)-windows-386.exe .
	@echo "✓  Built dist/$(APP)-windows-386.exe"

sign:
	@test -f "$(CERT_PATH)" || (echo "❌ CERT_PATH not set or file missing" && exit 1)
	osslsigncode sign \
		-pkcs12  "$(CERT_PATH)" \
		-pass    "$(CERT_PASSWORD)" \
		-n       "MoraLink GOst" \
		-i       "https://orbis.com.br" \
		-t       "http://timestamp.digicert.com" \
		-in      "$(FILE)" \
		-out     "$(FILE).signed"
	mv "$(FILE).signed" "$(FILE)"
	@echo "✓  Signed $(FILE)"

## Build installer and sign everything
release: windows installer
	$(MAKE) sign FILE=dist/$(APP)-windows-amd64.exe
	$(MAKE) sign FILE=dist/moralink-setup.exe
	@echo "✓  Release ready"
	
installer: windows windows32
	@echo "→  Building Windows installers..."
	cd build_assets && makensis installer.nsi
	cd build_assets && makensis installer32.nsi
	@echo "✓  Built dist/moralink-setup.exe and dist/moralink-setup-x86.exe"
	rm -rf ./clients_setups/*
	mkdir -p clients_setups/$(SHARK_TOKEN)
	find dist/ -maxdepth 1 -type f -exec mv {} clients_setups/$(SHARK_TOKEN) \;  

	

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