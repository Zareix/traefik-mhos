#!/bin/bash

OS="linux darwin windows"
ARCH="amd64 arm64"

for os in $OS; do
  for arch in $ARCH; do
    echo "Building for $os/$arch"
    GOOS=$os GOARCH=$arch go build -o "bin/traefik-mhos_$os-$arch"
  done
done
