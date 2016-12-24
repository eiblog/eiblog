#!/bin/bash

VERSION=`git describe --tags`
tar zcvf eiblog-$VERSION-linux-amd64.tar.gz conf static views eiblog

GOOS=windows GOARCH=amd64 go build && \
tar zcvf eiblog-$VERSION-windows-amd64.tar.gz conf static views eiblog
        
GOOS=darwin GOARCH=amd64 go build && \
tar zcvf eiblog-$VERSION-darwin-amd64.tar.gz conf static views eiblog
