.PHONY: all release darwin
.DEFAULT_GOAL := all

VERSION=v0.1.0
CWD=$(shell pwd)
GIT_EXEC=$(shell git --exec-path)

define build_release
	GOOS=darwin GOARCH=amd64 go build -o release/darwin-amd64/$(1) cmd/$(1)/$(1).go
	GOOS=linux GOARCH=386 go build -o release/linux-386/$(1) cmd/$(1)/$(1).go
	GOOS=linux GOARCH=amd64 go build -o release/linux-amd64/$(1) cmd/$(1)/$(1).go
endef

define darwin_install
	cp release/darwin-amd64/$(1) darwin/root/usr/local/clip/bin/$(1)
endef

release:
	rm -rf release
	mkdir -p release
	$(call build_release,clip)
	$(call build_release,clip-remote)
	cd release/darwin-amd64 && tar -zvcf ../clip-$(VERSION)-darwin-amd64.tar.gz *
	cd release/linux-386 && tar -zvcf ../clip-$(VERSION)-linux-386.tar.gz *
	cd release/linux-amd64 && tar -zvcf ../clip-$(VERSION)-linux-amd64.tar.gz *

install:
	go install github.com/thrawn01/clip/...
	ln -s $$GOPATH/bin/clip ${GIT_EXEC}/git-clip
	ln -s $$GOPATH/bin/clip-remote ${GIT_EXEC}/git-clip-remote

pkg:
	mkdir -p darwin/root/usr/local/clip/bin
	$(call darwin_install,clip)
	$(call darwin_install,clip-remote)
	pkgbuild --identifier org.thrawn01.clip --version $(VERSION) --scripts darwin/scripts --root darwin/root release/org.thrawn01.clip.pkg
	productbuild --distribution darwin/Distribution --package-path release/ release/clip$(VERSION)-darwin-amd64.pkg
