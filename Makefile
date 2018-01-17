.PHONY: version

all: version test build

version:
	git describe --tags > VERSION

test:
	go test -v ./...

build: version
	go build -o visit -ldflags "-X main.version=$(shell cat VERSION)"
