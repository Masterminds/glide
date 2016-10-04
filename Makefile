GLIDE_GO_EXECUTABLE ?= go
VERSION := $(shell git describe --tags)
DIST_DIRS := find * -type d -exec

build:
	${GLIDE_GO_EXECUTABLE} build -o glide -ldflags "-X main.version=${VERSION}" glide.go

install: build
	install -d ${DESTDIR}/usr/local/bin/
	install -m 755 ./glide ${DESTDIR}/usr/local/bin/glide

test:
	${GLIDE_GO_EXECUTABLE} test . ./gb ./path ./action ./tree ./util ./godep ./godep/strip ./gpm ./cfg ./dependency ./importer ./msg ./repo ./mirrors

integration-test:
	${GLIDE_GO_EXECUTABLE} build
	./glide up
	./glide install

clean:
	rm -f ./glide.test
	rm -f ./glide
	rm -rf ./dist

bootstrap-dist:
	${GLIDE_GO_EXECUTABLE} get -u github.com/franciscocpg/gox
	cd ${GOPATH}/src/github.com/franciscocpg/gox && git checkout dc50315fc7992f4fa34a4ee4bb3d60052eeb038e
	cd ${GOPATH}/src/github.com/franciscocpg/gox && ${GLIDE_GO_EXECUTABLE} install


build-all:
	gox -verbose \
	-ldflags "-X main.version=${VERSION}" \
	-os="linux darwin windows freebsd openbsd netbsd" \
	-arch="amd64 386 armv5 armv6 armv7 arm64" \
	-osarch="!darwin/arm64" \
	-output="dist/{{.OS}}-{{.Arch}}/{{.Dir}}" .

dist: build-all
	cd dist && \
	$(DIST_DIRS) cp ../LICENSE {} \; && \
	$(DIST_DIRS) cp ../README.md {} \; && \
	$(DIST_DIRS) tar -zcf glide-${VERSION}-{}.tar.gz {} \; && \
	$(DIST_DIRS) zip -r glide-${VERSION}-{}.zip {} \; && \
	cd ..


.PHONY: build test install clean bootstrap-dist build-all dist integration-test
