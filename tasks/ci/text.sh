#!/usr/bin/env bash

# Colors.
RED="\e[31m"
GREEN="\e[32m"
YELLOW="\e[33m"
RESET="\e[0m"

function fold_start {
    if [ "$TRAVIS" == "true" ]; then
        echo -en "travis_fold:start:$1\r"
        echo "\$ $2"
    fi
}

function fold_end {
    if [ "$TRAVIS" == "true" ]; then
        echo -en "travis_fold:end:$1\r"
    fi
}

function run_tests {
    go test "$1" -covermode=count -coverprofile=tmp.cov
    build_result=$?
    # Can't do race tests at the same time as coverage as it'll report
    # lots of false positives then..
    go test -race "$1"
    let build_result=$build_result+$?
    echo -ne "${YELLOW}=>${RESET} test $1 - "
    if [ "$build_result" == "0" ]; then
        echo -e "${GREEN}SUCCEEDED${RESET}"
    else
        echo -e "${RED}FAILED${RESET}"
    fi
}

function test_all {
    let a=0
    for pkg in $(go list "./$1/..."); do
        run_tests "$pkg"
        let a=$a+$build_result
        if [ "$build_result" == "0" ]; then
            sed 1d tmp.cov >> coverage.cov
        fi
    done
    build_result=$a
}

fold_start "get.cov" "get coverage tools"
go get golang.org/x/tools/cmd/cover
go get github.com/mattn/goveralls
go get github.com/axw/gocov/gocov
fold_end "get.cov"

echo "mode: count" > coverage.cov

ret=0

fold_start "test.all" "test all"
test_all "."
let ret=$ret+$build_result
fold_end "test.all"

if [ "$ret" == "0" ]; then
    fold_start "coveralls" "post to coveralls"
    "$(go env GOPATH | awk 'BEGIN{FS=":"} {print $1}')/bin/goveralls" -coverprofile=coverage.cov -service=travis-ci
    let ret=$ret+$?
    fold_end "coveralls"
fi

exit $ret
