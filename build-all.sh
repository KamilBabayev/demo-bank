#!/bin/bash

# Master Build Script for All Services
# Builds and pushes all Docker images to Docker Hub

set -e  # Exit on error

echo "=========================================="
echo "Building All Microservices"
echo "=========================================="
echo ""

# Configuration
DOCKER_USERNAME="kamilbabayev"

# Build API Gateway
echo "📦 Building API Gateway..."
cd /Users/kamil/git/demo-bank/api-gateway
./build.sh
echo ""

# Build User Service
if [ -f "/Users/kamil/git/demo-bank/user/Dockerfile" ]; then
    echo "📦 Building User Service..."
    cd /Users/kamil/git/demo-bank/user
    ./build.sh
    echo ""
fi

# Build Account Service
if [ -f "/Users/kamil/git/demo-bank/account/Dockerfile" ]; then
    echo "📦 Building Account Service..."
    cd /Users/kamil/git/demo-bank/account
    ./build.sh
    echo ""
fi

# Build Payment Service
if [ -f "/Users/kamil/git/demo-bank/payment/Dockerfile" ]; then
    echo "📦 Building Payment Service..."
    cd /Users/kamil/git/demo-bank/payment
    ./build.sh
    echo ""
fi

# Build Transfer Service
if [ -f "/Users/kamil/git/demo-bank/transfer/Dockerfile" ]; then
    echo "📦 Building Transfer Service..."
    cd /Users/kamil/git/demo-bank/transfer
    ./build.sh
    echo ""
fi

# Build Notification Service
if [ -f "/Users/kamil/git/demo-bank/notification/Dockerfile" ]; then
    echo "📦 Building Notification Service..."
    cd /Users/kamil/git/demo-bank/notification
    ./build.sh
    echo ""
fi

echo "=========================================="
echo "✅ All builds complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  1. Deploy infrastructure: kubectl apply -f infra/"
echo "  2. Deploy services: kubectl apply -f api-gateway/deploy/"
echo "  3. Check status: kubectl get pods --all-namespaces"
