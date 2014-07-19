VERSION := $(shell git describe --tags)

build:
	go build -o glide -ldflags "-X main.version ${VERSION}" glide.go

install: build
	install -d ${DESTDIR}/usr/local/bin/
	install -m 755 ./glide ${DESTDIR}/usr/local/bin/glide

test:
	go test ./...

clean:
	rm -f ./glide.test
	rm -f ./glide

bootstrap:
	mkdir ./_vendor
	GOPATH=${PWD}/_vendor go get github.com/Masterminds/cookoo
	GOPATH=${PWD}/_vendor go get github.com/kylelemons/go-gypsy/yaml
	ln -s ${PWD} _vendor/src/github.com/Masterminds/glide
	GOPATH=${PWD}/_vendor go build -o glide -ldflags "-X main.version ${VERSION}" glide.go

.PHONY: build test install clean
