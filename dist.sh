#!/bin/bash

# GOFLAGS='-ldflags="-s -w"'
version=`git describe --tags`
arch=$(go env GOARCH)

for os in linux darwin windows; do
    echo "... building $version for $os/$arch"
    TARGET="eiblog-$version.$os-$arch"
    GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build
    tar czvf $TARGET.tar.gz conf static views eiblog
    rm eiblog
done
