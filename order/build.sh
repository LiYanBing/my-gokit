#!/bin/sh
IMAGE_TAG=v0.0.1
SERVICE_NAME=order
IMAGE_NAME=hub.docker.com${SERVICE_NAME}

set -o errexit
GOOS=linux GOARCH=amd64 go build -i -o _bin/${SERVICE_NAME}
docker build . -t ${IMAGE_NAME}:${IMAGE_TAG}
docker push ${IMAGE_NAME}:${IMAGE_TAG}
rm -f _bin/${SERVICE_NAME}