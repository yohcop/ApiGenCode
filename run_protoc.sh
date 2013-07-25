#!/bin/bash

protoc \
    -Ithird_party/src/code.google.com/p/protobuf-git/src \
    -Itests/ \
    -Isrc/ \
    --plugin=protoc-gen-pbform=bin/protoc-gen-pbform \
    --pbform_out=/tmp/foo \
    $*
