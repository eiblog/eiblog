#!/usr/bin/env sh

go build -tags prod -ldflags '-extldflags "-static"' -o "./cmd/$1/backend" "./cmd/$1"
