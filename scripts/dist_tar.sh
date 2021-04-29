#!/usr/bin/env sh

set -e

_tag="$1"
_arch=$(go env GOARCH)

# prepare dir ./bin
mkdir -p ./bin

for file in pkg/core/*; do
  app="$(basename $file)";
  # tar platform
  for os in linux darwin windows; do
    _target="$app-$_tag.$os-$_arch.tar.gz"
    CGO_ENABLED=0 GOOS=$os GOARCH=$_arch \
      go build -o bin/backend "./cmd/$app"
    if [ "$app" = "eiblog" ]; then
      tar czf $_target conf website assets bin/backend
    else
      tar czf $_target conf bin/backend
    fi
  done
done

# clean dir ./bin
rm -rf ./bin
