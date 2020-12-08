#!/bin/sh
IMAGE_TAG=latest
SERVICE_NAME=health
IMAGE_REGISTRY=hub.docker.com/sobe
IMAGE_NAME=${IMAGE_REGISTRY}/${SERVICE_NAME}

set -o errexit
GOOS=linux GOARCH=amd64 go build -i -o _bin/${SERVICE_NAME}
docker build . -t ${IMAGE_NAME}:${IMAGE_TAG}
docker push ${IMAGE_NAME}:${IMAGE_TAG}
rm -f _bin/${SERVICE_NAME}