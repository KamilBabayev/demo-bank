# API Gateway - Build & Deploy Guide

## Overview

This service uses **Minikube's Docker daemon** to build images locally. No need to push to Docker Hub!

## Build & Deploy Steps

### 1. Build Docker Image in Minikube

```bash
# Navigate to api-gateway directory
cd /Users/kamil/git/demo-bank/api-gateway

# Run the build script
./build.sh
```

**What this does:**
- Points your Docker CLI to Minikube's Docker daemon
- Builds the image directly inside Minikube
- Image is tagged as `api-gateway:latest`

### 2. Deploy to Kubernetes

```bash
# Create namespace
kubectl apply -f deploy/namespace.yaml

# Deploy API Gateway
kubectl apply -f deploy/deployment.yaml
kubectl apply -f deploy/service.yaml
```

### 3. Verify Deployment

```bash
# Check if pod is running
kubectl get pods -n api-gateway

# Check service
kubectl get svc -n api-gateway

# View logs
kubectl logs -n api-gateway -l app=api-gateway -f
```

### 4. Access the Service

```bash
# Get Minikube IP
minikube ip

# API Gateway is exposed on NodePort 30080
curl http://$(minikube ip):30080/health
```

## Manual Build (Alternative)

If you prefer manual steps:

```bash
# 1. Point Docker to Minikube
eval $(minikube docker-env)

# 2. Build image
docker build -t api-gateway:latest .

# 3. Verify image exists
docker images | grep api-gateway

# 4. Deploy to Kubernetes
kubectl apply -f deploy/
```

## Rebuild After Code Changes

```bash
# Just run the build script again
./build.sh

# Then restart the deployment to use new image
kubectl rollout restart deployment/api-gateway -n api-gateway
```

## Troubleshooting

### Image not found
```bash
# Make sure you're using Minikube's Docker
eval $(minikube docker-env)

# Rebuild
docker build -t api-gateway:latest .
```

### Pod not starting
```bash
# Check pod status
kubectl describe pod -n api-gateway <pod-name>

# Check logs
kubectl logs -n api-gateway <pod-name>
```

### Wrong Docker daemon
```bash
# Reset to Minikube's Docker
eval $(minikube docker-env)

# Verify you're using Minikube's Docker
docker info | grep -i "Operating System"
# Should show something related to Minikube
```

## Key Configuration

In `deploy/deployment.yaml`:
```yaml
imagePullPolicy: Never  # Use local image, don't pull from registry
```

This is crucial! It tells Kubernetes to use the local image built in Minikube.
