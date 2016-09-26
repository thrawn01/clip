.PHONY: all release darwin
.DEFAULT_GOAL := all

VERSION=v0.1.0
CWD=$(shell pwd)

define build_release
	GOOS=darwin GOARCH=amd64 go build -o release/$(1)-$(VERSION)-darwin-amd64 $(1).go
	GOOS=linux GOARCH=386 go build -o release/$(1)-$(VERSION)-linux-386 $(1).go
	GOOS=linux GOARCH=amd64 go build -o release/$(1)-$(VERSION)-linux-amd64 $(1).go
endef

define install
	cp release/$(1)-$(VERSION)-darwin-amd64 darwin/root/usr/local/clip/bin/$(1)
endef

git-clip: cmd/clip/clip.go
	go build -o git-clip cmd/clip/clip.go

git-clip-remote: cmd/clip-remote/clip-remote.go
	go build -o git-clip-remote cmd/clip-remote/clip-remote.go

all: git-clip git-clip-remote

clean:
	rm git-clip git-clip-remote

release:
	rm -rf release
	mkdir -p release
	$(call build_release,git-clip)
	$(call build_release,git-clip-remote)

install:
	cd `git --exec-path`;\
		sudo ln -s $(CWD)/git-clip .;\
		sudo ln -s $(CWD)/git-clip-remote .

pkg:
	mkdir -p darwin/root/usr/local/clip/bin
	$(call install,clip)
	$(call install,clip-remote)
	pkgbuild --identifier org.thrawn01.clip --version $(VERSION) --scripts darwin/scripts --root darwin/root release/org.thrawn01.clip.pkg
	productbuild --distribution darwin/Distribution --package-path release/ release/clip$(VERSION)-darwin-amd64.pkg
