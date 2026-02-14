#!/bin/bash
set -e

DOCKER_USERNAME="kamilbabayev"
IMAGE_NAME="${DOCKER_USERNAME}/demo-bank:frontend"

echo "Building frontend Docker image..."
docker build -t ${IMAGE_NAME} .

echo "Pushing image to Docker Hub..."
docker push ${IMAGE_NAME}

echo ""
echo "Build complete! Image: ${IMAGE_NAME}"
echo ""
echo "Next steps:"
echo "  kubectl apply -f deploy/namespace.yaml"
echo "  kubectl apply -f deploy/"
