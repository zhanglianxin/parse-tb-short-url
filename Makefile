SHELL := /bin/bash
PLATFORM := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin

GO_PACKAGE := github.com/zhanglianxin/parse-tb-short-url
CROSS_TARGETS := linux/amd64 darwin/amd64 windows/386 windows/amd64

default: build cross gen-sha1

get-deps:
	dep ensure

cp-config:
	cp config_example.toml config.toml

build:
	go fmt ./...
	@#go build

clean:
	rm -fr data/*

cross:
	gox -osarch="$(CROSS_TARGETS)" $(GO_PACKAGE)
	@$(MAKE) gen-sha1

rm-sha1:
	@rm -f parse-tb-short-url_*.sha1

gen-sha1: rm-sha1
	@$$(for f in $$(find parse-tb-short-url_* -type f); do shasum $$f > $$f.sha1; done)
