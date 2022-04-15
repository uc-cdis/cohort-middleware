#/bin/bash

IMAGE_TAG=$1

echo "=== Running integration tests ====="
echo "..."

docker run --rm \
    -v $PWD/.coverage_out:/coverage_out \
    $IMAGE_TAG \
    sh -c "
    go test ./... -v -coverpkg=./... -cover -coverprofile=/coverage_out/coverage.out && \
    ls -la /coverage_out/coverage.out && \
    go tool cover -html=/coverage_out/coverage.out -o /coverage_out/coverage.html && \
    ls -la /coverage_out/coverage.html"

if [ "$2" ]; then
    echo "==== Generating coveralls report....."
    COVERALLS_TOKEN=$2

    docker run --rm \
        -v $PWD/.coverage_out:/coverage_out \
        $IMAGE_TAG \
        sh -c "
        ls -la /coverage_out/coverage.out && \
        go get github.com/mattn/goveralls \
        goveralls --coverprofile=/coverage_out/coverage.out --service=github --repotoken $COVERALLS_TOKEN"
fi
