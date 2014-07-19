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

.PHONY: build test clean install
