#!/usr/bin/env sh

_registry="$1"
_tag="$2"
_platform="linux/amd64,linux/arm64,linux/386"

if [ -z "$_registry" ] || [ -z "$_tag" ]; then
  echo "Please specify image repository and tag."
  exit 0;
fi

# create and use builder
docker buildx inspect builder >/dev/null 2>&1
if [ "$?" != "0" ]; then
  docker buildx create --use --name builder
fi

# prepare dir ./bin
mkdir -p ./bin
# build demo app
for file in pkg/core/*; do
  app="$(basename $file)";
  go build -tags prod -ldflags '-extldflags "-static"' -o bin/backend "./cmd/$app"
  # docker image
  docker buildx build --platform "$_platform" \
    -f "build/package/$app.Dockerfile" \
    -t "$_registry/$app:latest" \
    -t "$_registry/$app:$_tag" \
    --push .
done

# clean dir ./bin
rm -rf ./bin
