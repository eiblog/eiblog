#!/bin/bash

# GOFLAGS='-ldflags="-s -w"'
version=`git describe --tags`
arch=$(go env GOARCH)

for os in linux darwin windows; do
    echo "... building $version for $os/$arch"
    TARGET="eiblog-$version.$os-$arch"
    GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build
    if [ "$os" == "windows" ]; then
        tar czvf $TARGET.tar.gz conf static views eiblog.exe
        rm eiblog.exe
    else
        tar czvf $TARGET.tar.gz conf static views eiblog
        rm eiblog
    fi
done
