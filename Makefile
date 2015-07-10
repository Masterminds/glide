VERSION := $(shell git describe --tags)

build:
	GOPATH=${PWD}/vendor go build -o glide -ldflags "-X main.version ${VERSION}" glide.go

install: build
	install -d ${DESTDIR}/usr/local/bin/
	install -m 755 ./glide ${DESTDIR}/usr/local/bin/glide

test:
	go test ./...

clean:
	rm -f ./glide.test
	rm -f ./glide

bootstrap:
	mkdir ./vendor
	GOPATH=${PWD}/vendor go get github.com/Masterminds/cookoo
	GOPATH=${PWD}/vendor go get github.com/kylelemons/go-gypsy/yaml
	GOPATH=${PWD}/vendor go get github.com/codegangsta/cli
	ln -s ${PWD} vendor/src/github.com/Masterminds/glide
	GOPATH=${PWD}/vendor go build -o glide -ldflags "-X main.version ${VERSION}" glide.go

.PHONY: build test install clean
