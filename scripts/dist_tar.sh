#!/usr/bin/env sh

set -e

_tag="$1"
_arch=$(go env GOARCH)

for file in pkg/core/*; do
  app="$(basename $file)";
  # tar platform
  for os in linux darwin windows; do
    _target="$app-$_tag.$os-$_arch.tar.gz"
    CGO_ENABLED=0 GOOS=$os GOARCH=$_arch \
      go build -tags prod -o backend "./cmd/$app"
    if [ "$app" = "eiblog" ]; then
      tar czf $_target conf website assets backend
    else
      tar czf $_target conf backend
    fi
  done
done
