#/bin/bash

IMAGE_TAG=$1

docker build . --file ./tests/build_test_image/Dockerfile --tag $IMAGE_TAG
