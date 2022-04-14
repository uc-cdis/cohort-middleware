#/bin/bash

IMAGE_TAG=$1

echo "=== Running integration tests ====="
echo "..."

docker run --rm \
    $IMAGE_TAG \
    go test ./... -v -coverpkg=./...
