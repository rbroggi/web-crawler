#!/bin/sh

echo "tests"
/go/src/app/crawler_lib_test -test.v
echo "command line test"
/go/src/app/cmd/crawler -url=${CRAWLER_BASE_URL}/index.html