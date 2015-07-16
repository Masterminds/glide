VERSION := $(shell git describe --tags)

build:
	go build -o glide -ldflags "-X main.version=${VERSION}" glide.go

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
	cd ./vendor
	git clone https://github.com/Masterminds/cookoo github.com/Masterminds/cookoo
	git clone https://github.com/kylelemons/go-gypsy github.com/kylelemons/go-gypsy
	git clone https://github.com/codegangsta/cli github.com/codegangsta/cli
	go get golang.org/x/tools/go/vcs

.PHONY: build test install clean
