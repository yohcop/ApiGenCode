export GOPATH=$GOPATH:${PWD}/genfiles
./build.sh && \
    bin/gen --schema=api1.json --out=genfiles/src/api1 --gen_go_pkg=api1 && \
    go test src/tests/api1_test.go
