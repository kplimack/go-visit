.PHONY: version

all: version test build

version:
	git describe --tags > VERSION

test:
	go test -v ./...

build: version
	CGO_ENABLED=0 GOOS=linux go build -o build/visit -ldflags "-X main.version=$(shell cat VERSION)"

docker: version
	docker build -t partkyle/go-visit:$(shell cat VERSION) .

docker-push: docker
	docker push partkyle/go-visit:$(shell cat VERSION)

clean:
	rm -rf build
