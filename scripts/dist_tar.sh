#!/usr/bin/env sh

set -e

_tag="$1"
_arch=$(go env GOARCH)

for file in cmd/*; do
  # Skip if not a directory
  if [ ! -d "$file" ]; then
    continue
  fi

  app="$(basename $file)";
  # tar platform
  for os in linux darwin windows; do
    _target="$app-$_tag.$os-$_arch.tar.gz"
    GOOS=$os GOARCH=$_arch scripts/run_build.sh $app

    # Create tar with flattened structure using -C parameter
    tar czf "$_target" \
      CHANGELOG.md LICENSE README.md \
      -C "./cmd/$app" etc backend
  done
done
