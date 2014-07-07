#!/bin/bash

protoc \
    -Ithird_party/src/code.google.com/p/protobuf-git/src \
    -Itests/ \
    -Isrc/ \
    --plugin=protoc-gen-pbform=bin/protoc-gen-pbform \
    --pbform_out=tpl_path=src/protoc-gen-pbform,gen_html_form=true,gen_go_services=true,override_js=true:/tmp/foo \
    $*
