#!/bin/bash
#
# Must run get.sh before using.
#
# Example build/usage:
# ./build.sh && bin/gen --schema=api1.json

protoc \
    -Ithird_party/src/code.google.com/p/protobuf-git/src \
    -Isrc \
    --plugin=protoc-gen-go=third_party/bin/protoc-gen-go \
    --go_out=genfiles/src \
    src/pbform/*.proto

protoc \
    -Ithird_party/src/code.google.com/p/protobuf-git/src \
    -Isrc \
    --plugin=protoc-gen-go=third_party/bin/protoc-gen-go \
    --go_out=genfiles/src \
    third_party/src/code.google.com/p/protobuf-git/src/google/protobuf/descriptor.proto

protoc \
    -Ithird_party/src/code.google.com/p/protobuf-git/src \
    -Isrc \
    --plugin=protoc-gen-go=third_party/bin/protoc-gen-go \
    --go_out=genfiles/src \
    third_party/src/code.google.com/p/protobuf-git/src/google/protobuf/compiler/plugin.proto

P=`pwd $0`
export GOPATH=$P/third_party:$P/genfiles:$P

go install gen protoc-gen-pbform

