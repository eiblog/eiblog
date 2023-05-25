#!/usr/bin/env sh

go build -tags prod -ldflags '-extldflags "-static"' -o bin/backend "./cmd/$1"
