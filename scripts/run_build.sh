#!/usr/bin/env sh

go build -ldflags '-extldflags "-static"' -o ./cmd/$1/backend ./cmd/$1
