PKG=github.com/parkr/instapaper-archive
REV ?=$(shell git rev-parse HEAD)
TAG=ghcr.io/parkr/instapaper-archive:$(REV)

all: fmt build test

fmt:
	go fmt $(PKG)/...

build:
	go build $(PKG)

test:
	go test $(PKG)

docker-build:
	docker build -t $(TAG) .

docker-test: docker-build
	docker run --rm $(TAG) -h

docker-release: docker-build
	docker push $(TAG)

dive: docker-build
	dive $(TAG)

env:
	@echo $(TAG)
