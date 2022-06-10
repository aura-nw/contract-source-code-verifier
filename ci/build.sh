#!/bin/sh
set -xe

#Login to registry
echo $GITHUB_PASSWORD | docker login ghcr.io -u $GITHUB_USERNAME --password-stdin
#Build and push image
export DOCKER_HOST="tcp://localhost:2375"
docker build -t ${CONTAINER_RELEASE_IMAGE} -f Dockerfile .
docker push ${CONTAINER_RELEASE_IMAGE}