#!/bin/bash
#
# Must run get.sh before using.
#
# Example build/usage:
# ./build.sh && bin/gen --schema=api1.json

P=`pwd $0`
export GOPATH=$P/third_party:$P/genfiles:$P

go install gen

