#!/usr/bin/env sh

set -e

_tag="$1"
_arch=$(go env GOARCH)

for file in cmd/*; do
  app="$(basename $file)";
  # tar platform
  for os in linux darwin windows; do
    _target="$app-$_tag.$os-$_arch.tar.gz"
    GOOS=$os GOARCH=$_arch scripts/run_build.sh $app
    tar czf $_target ./cmd/$app/etc ./cmd/$app/backend
  done
done
