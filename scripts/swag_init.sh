#!/usr/bin/env sh

for file in pkg/core/*; do
  if test -d $file; then
    cd $file && swag init -g api.go;
    cd -;
  fi
done

