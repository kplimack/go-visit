.PHONY: version

all: version test build

version:
	git describe --tags > VERSION

test:
	go test -v ./...

build: version
	go build -o build/visit -ldflags "-X main.version=$(shell cat VERSION)"

clean:
	rm -rf build
