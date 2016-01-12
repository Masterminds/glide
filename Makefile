VERSION := $(shell git describe --tags)
DIST_DIRS := find * -type d -exec

build:
	go build -o glide -ldflags "-X main.version=${VERSION}" glide.go

install: build
	install -d ${DESTDIR}/usr/local/bin/
	install -m 755 ./glide ${DESTDIR}/usr/local/bin/glide

test:
	go test . ./cmd ./gb

clean:
	rm -f ./glide.test
	rm -f ./glide
	rm -rf ./dist

bootstrap:
	rm -rf ./vendor
	mkdir ./vendor
	
	git clone https://github.com/Masterminds/cookoo vendor/github.com/Masterminds/cookoo
	rm -rf vendor/github.com/Masterminds/cookoo/.git
	git clone https://github.com/Masterminds/vcs vendor/github.com/Masterminds/vcs
	rm -rf vendor/github.com/Masterminds/vcs/.git
	git clone https://gopkg.in/yaml.v2 vendor/gopkg.in/yaml.v2
	rm -rf vendor/gopkg.in/yaml.v2/.git
	git clone https://github.com/codegangsta/cli vendor/github.com/codegangsta/cli
	rm -rf vendor/github.com/codegangsta/cli/.git
	git clone https://github.com/Masterminds/semver vendor/github.com/Masterminds/semver
	rm -rf vendor/github.com/Masterminds/semver/.git

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
