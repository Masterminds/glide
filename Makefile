VERSION := $(shell git describe --tags)
DIST_DIRS := find * -type d -exec

build:
	go build -o glide -ldflags "-X main.version=${VERSION}" glide.go

install: build
	install -d ${DESTDIR}/usr/local/bin/
	install -m 755 ./glide ${DESTDIR}/usr/local/bin/glide

test:
	go test . ./cmd ./gb ./path ./action

clean:
	rm -f ./glide.test
	rm -f ./glide
	rm -rf ./dist

bootstrap:
	mkdir ./vendor
	git clone https://github.com/Masterminds/cookoo vendor/github.com/Masterminds/cookoo
	git clone https://github.com/Masterminds/vcs vendor/github.com/Masterminds/vcs
	git clone https://gopkg.in/yaml.v2 vendor/gopkg.in/yaml.v2
	git clone https://github.com/codegangsta/cli vendor/github.com/codegangsta/cli
	git clone https://github.com/Masterminds/semver vendor/github.com/Masterminds/semver

bootstrap-dist:
	go get -u github.com/mitchellh/gox

build-all:
	gox -verbose \
	-ldflags "-X main.version=${VERSION}" \
	-os="linux darwin windows " \
	-arch="amd64 386" \
	-output="dist/{{.OS}}-{{.Arch}}/{{.Dir}}" .

dist: build-all
	cd dist && \
	$(DIST_DIRS) cp ../LICENSE {} \; && \
	$(DIST_DIRS) cp ../README.md {} \; && \
	$(DIST_DIRS) tar -zcf glide-${VERSION}-{}.tar.gz {} \; && \
	$(DIST_DIRS) zip -r glide-${VERSION}-{}.zip {} \; && \
	cd ..


.PHONY: build test install clean bootstrap bootstrap-dist build-all dist
