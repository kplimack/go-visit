.PHONY: version test build docker docker-push docker-push-latest clean

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

docker-push-latest: docker docker-push
	docker tag partkyle/go-visit:$(shell cat VERSION) partkyle/go-visit:latest
	docker push partkyle/go-visit:latest

clean:
	rm -rf build
