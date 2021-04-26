#!/usr/bin/env sh

set -e

for file in */*.proto; do
  if test -f $file; then
    protoc --go_out=. --go_opt=paths=source_relative \
      --go-grpc_out=. --go-grpc_opt=paths=source_relative \
      $file;
  fi
done
