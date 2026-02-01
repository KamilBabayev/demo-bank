#!/bin/bash

# Payment Service Build Script for Docker Hub

set -e  # Exit on error

echo "====================================="
echo "Building Payment Docker Image"
echo "====================================="

# Configuration
DOCKER_USERNAME="kamilbabayev"
IMAGE_NAME="${DOCKER_USERNAME}/demo-bank:payment"

# Step 1: Build the Docker image
echo "Building Docker image..."
docker build -t ${IMAGE_NAME} .

# Step 2: Push to Docker Hub
echo ""
echo "Pushing to Docker Hub..."
docker push ${IMAGE_NAME}

# Step 3: Verify
echo ""
echo "✅ Build complete! Image pushed to Docker Hub"
echo ""
echo "Image: ${IMAGE_NAME}"
echo ""
echo "Next steps:"
echo "  kubectl apply -f deploy/namespace.yaml"
echo "  kubectl apply -f deploy/"
