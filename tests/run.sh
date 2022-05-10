#/bin/bash

echo "=== Running integration tests ====="
echo "..."

COVERAGE_OUT=$PWD/.coverage_out
mkdir -p $COVERAGE_OUT
go test ./... -v -coverpkg=../... -race -covermode atomic -coverprofile=$COVERAGE_OUT/coverage.out
go tool cover -html=$COVERAGE_OUT/coverage.out -o $COVERAGE_OUT/coverage.html

# open coverage report in browser:
open $COVERAGE_OUT/coverage.html
