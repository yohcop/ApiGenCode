#!/bin/sh

P=`pwd $0`
export GOPATH=$P/third_party:$P

go install \
  code.google.com/p/goprotobuf/protoc-gen-go \
  code.google.com/p/goprotobuf/proto

mkdir -p genfiles/src
