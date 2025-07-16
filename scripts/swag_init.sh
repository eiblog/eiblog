#!/usr/bin/env sh

for file in cmd/*; do
  if test -d $file; then
    cd $file && swag init -g main.go;
    cd -;
  fi
done

