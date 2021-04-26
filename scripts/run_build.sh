#!/usr/bin/env sh

set -e

_registry="$1"
_tag="$2"

if [ -z "$_registry" ] || [ -z "$_tag" ]; then
  echo "Please specify image repository and tag."
  exit 0;
fi

# prepare dir ./bin
mkdir -p ./bin

# build demo app
for file in pkg/core/*; do
  app="$(basename $file)";
  GOOS=linux GOARCH=amd64 go build -o bin/backend "./cmd/$app"
  docker build -f "build/package/$app.Dockerfile" -t "$_registry/$app:$_tag" .
done

# clean dir ./bin
rm -rf ./bin
