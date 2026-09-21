SHELL := /bin/bash

VERSION=$(shell git describe --abbrev=0 --always)
LDFLAGS = -ldflags "-X main.version=${VERSION}"
GCFLAGS = -gcflags="all=-N -l"
PACKAGE=./cmd/ufo
EXECUTABLE=ufo
WINDOWS=$(EXECUTABLE)_windows_amd64.exe
WINDOWS_ARM64=$(EXECUTABLE)_windows_arm64.exe
LINUX=$(EXECUTABLE)_linux_amd64
LINUX_ARM64=$(EXECUTABLE)_linux_arm64
DARWIN_AMD64=$(EXECUTABLE)_darwin_amd64
DARWIN_ARM64=$(EXECUTABLE)_darwin_arm64
ALL=$(WINDOWS) $(WINDOWS_ARM64) $(LINUX) $(LINUX_ARM64) $(DARWIN_AMD64) $(DARWIN_ARM64)
ZIPS=$(addsuffix _$(VERSION).zip,$(basename $(ALL)))
RELEASE_ASSETS=$(ZIPS) SHA256SUMS-$(VERSION)
# ufo is licensed under the LGPL, whose text has to accompany the binary.
# The source of each release is the tag it is built from, which GitHub offers
# as an archive on the release page.
ZIP_EXTRAS=LICENSE README.md

default:
	go build ${LDFLAGS} -o $(EXECUTABLE) $(PACKAGE)

install:
	go install ${LDFLAGS} $(PACKAGE)

install-debug:
	go install ${LDFLAGS} ${GCFLAGS} $(PACKAGE)

$(WINDOWS):
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -v -o $(WINDOWS) ${LDFLAGS} $(PACKAGE)

$(WINDOWS_ARM64):
	env CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -v -o $(WINDOWS_ARM64) ${LDFLAGS} $(PACKAGE)

$(LINUX):
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o $(LINUX) ${LDFLAGS} $(PACKAGE)

$(LINUX_ARM64):
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -v -o $(LINUX_ARM64) ${LDFLAGS} $(PACKAGE)

$(DARWIN_AMD64):
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -v -o $(DARWIN_AMD64) ${LDFLAGS} $(PACKAGE)
	rcodesign sign --for-notarization --pem-file <(pass OctoberSwimmer/codesign/combined) $@

$(DARWIN_ARM64):
	env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -v -o $(DARWIN_ARM64) ${LDFLAGS} $(PACKAGE)
	rcodesign sign --for-notarization --pem-file <(pass OctoberSwimmer/codesign/combined) $@

$(basename $(WINDOWS))_$(VERSION).zip: $(WINDOWS)
	@rm -f $@
	zip $@ $< $(ZIP_EXTRAS)
	7za rn $@ $< $(EXECUTABLE)$(suffix $<)

$(basename $(WINDOWS_ARM64))_$(VERSION).zip: $(WINDOWS_ARM64)
	@rm -f $@
	zip $@ $< $(ZIP_EXTRAS)
	7za rn $@ $< $(EXECUTABLE)$(suffix $<)

$(basename $(DARWIN_AMD64))_$(VERSION).zip: $(DARWIN_AMD64)
	@rm -f $@
	zip $@ $< $(ZIP_EXTRAS)
	7za rn $@ $< $(EXECUTABLE)
	rcodesign notary-submit --api-key-file <(pass OctoberSwimmer/codesign/api-key) $@

$(basename $(DARWIN_ARM64))_$(VERSION).zip: $(DARWIN_ARM64)
	@rm -f $@
	zip $@ $< $(ZIP_EXTRAS)
	7za rn $@ $< $(EXECUTABLE)
	rcodesign notary-submit --api-key-file <(pass OctoberSwimmer/codesign/api-key) $@

%_$(VERSION).zip: %
	@rm -f $@
	zip $@ $< $(ZIP_EXTRAS)
	7za rn $@ $< $(EXECUTABLE)

dist: test $(ZIPS)

checksum: dist
	shasum -a 256 $(ZIPS) > SHA256SUMS-$(VERSION)

release: checksum
	@if ! command -v gh >/dev/null 2>&1; then \
		echo "gh CLI is required for 'make release'."; \
		exit 1; \
	fi
	@if ! git rev-parse --verify "refs/tags/$(VERSION)" >/dev/null 2>&1; then \
		echo "Tag '$(VERSION)' does not exist. Create the tag before running 'make release'."; \
		exit 1; \
	fi
	@if [ "$$(git describe --exact-match --tags HEAD 2>/dev/null)" != "$(VERSION)" ]; then \
		echo "HEAD is not exactly at tag '$(VERSION)'. Check out the tag before running 'make release'."; \
		exit 1; \
	fi
	git push octoberswimmer "$(VERSION)"
	gh release create "$(VERSION)" --title "ufo $(VERSION)" --notes-from-tag --verify-tag $(RELEASE_ASSETS)
	brew bump-cask-pr ufo --version $(VERSION:v%=%)

fmt:
	go fmt ./...

test:
	test -z "$$(gofmt -l .)"
	go vet ./...
	go test ./...
	go test -race ./...

clean:
	-rm -f $(EXECUTABLE) $(EXECUTABLE)_* *.zip SHA256SUMS-*

.PHONY: default install install-debug dist clean checksum release fmt test tag

tag:
	@echo "Creating next tag..."
	@bash -c ' \
	PREV_TAG=$$(git tag --sort=-version:refname | head -1); \
	if [ -z "$$PREV_TAG" ]; then \
		NEXT_TAG="v0.0.1"; \
	else \
		VERSION=$$(echo $$PREV_TAG | sed "s/v//"); \
		MAJOR=$$(echo $$VERSION | cut -d. -f1); \
		MINOR=$$(echo $$VERSION | cut -d. -f2); \
		PATCH=$$(echo $$VERSION | cut -d. -f3); \
		NEXT_MINOR=$$((MINOR + 1)); \
		NEXT_TAG="v$$MAJOR.$$NEXT_MINOR.0"; \
	fi; \
	echo "Previous tag: $$PREV_TAG"; \
	echo "Next tag: $$NEXT_TAG"; \
	echo ""; \
	echo "Changelog:"; \
	if [ -z "$$PREV_TAG" ]; then RANGE=HEAD; else RANGE=$$PREV_TAG..; fi; \
	git changelog $$RANGE; \
	echo ""; \
	read -p "Create tag $$NEXT_TAG? [y/N] " -n 1 -r; \
	echo ""; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		CHANGELOG=$$(git changelog $$RANGE); \
		git tag -a $$NEXT_TAG -m "Version $$NEXT_TAG" -m "" -m "$$CHANGELOG"; \
		echo "Tag $$NEXT_TAG created successfully"; \
		echo "Push with: git push octoberswimmer $$NEXT_TAG"; \
	else \
		echo "Tag creation cancelled"; \
	fi'
