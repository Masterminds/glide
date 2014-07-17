#!/bin/bash

if [[ -e _vendor ]]; then
  echo "Cowardly refusing to continue. _vendor exists already."
  exit 1
fi

GOPATH=$PWD/_vendor

mkdir -p _vendor/src/github.com/Masterminds

go get github.com/Masterminds/cookoo
go get github.com/kylelemons/go-gypsy/yaml

ln -s . _vendor/src/github.com/Masterminds/glide

echo "Dependencies prepared. Building ./glide."
go build
