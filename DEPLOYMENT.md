# Demo Bank - Deployment Guide

## Quick Start

### Prerequisites
- Minikube running: `minikube start`
- Docker installed
- kubectl configured

### Deploy Everything

```bash
# 1. Deploy Infrastructure (databases + Kafka)
cd k8s-infra
kubectl apply -f .

# 2. Build all service images
cd ..
./build-all.sh

# 3. Deploy all services
kubectl apply -f api-gateway/deploy/
kubectl apply -f user-service/deploy/
kubectl apply -f account-service/deploy/
kubectl apply -f payment-service/deploy/
kubectl apply -f transfer-service/deploy/

# 4. Check everything is running
kubectl get pods --all-namespaces
```

### Access API Gateway

```bash
# Get Minikube IP
minikube ip

# Access API Gateway (exposed on NodePort 30080)
curl http://$(minikube ip):30080/health
```

---

## Namespace Architecture

Each component runs in its own namespace:

```
Namespaces:
├── infrastructure          (Databases + Kafka)
├── api-gateway            (API Gateway)
├── user-service           (User Service)
├── account-service        (Account Service)
├── payment-service        (Payment Service)
└── transfer-service       (Transfer Service)
```

## Deploy Everything

### Step 1: Deploy Infrastructure
```bash
cd k8s-infra
kubectl apply -f namespace.yaml
kubectl apply -f .
kubectl get pods -n infrastructure
```

### Step 2: Deploy Services

Deploy each service in its namespace:

```bash
# API Gateway
kubectl apply -f api-gateway/deploy/namespace.yaml
kubectl apply -f api-gateway/deploy/

# User Service
kubectl apply -f user-service/deploy/namespace.yaml
kubectl apply -f user-service/deploy/

# Account Service
kubectl apply -f account-service/deploy/namespace.yaml
kubectl apply -f account-service/deploy/

# Payment Service
kubectl apply -f payment-service/deploy/namespace.yaml
kubectl apply -f payment-service/deploy/

# Transfer Service
kubectl apply -f transfer-service/deploy/namespace.yaml
kubectl apply -f transfer-service/deploy/
```

### Step 3: Verify All Namespaces
```bash
# List all namespaces
kubectl get namespaces

# Check pods in each namespace
kubectl get pods -n infrastructure
kubectl get pods -n api-gateway
kubectl get pods -n user-service
kubectl get pods -n account-service
kubectl get pods -n payment-service
kubectl get pods -n transfer-service
```

### Step 4: Access API Gateway
```bash
# Get Minikube IP
minikube ip

# API Gateway is exposed on NodePort 30080
# Access: http://<minikube-ip>:30080
```

## Cross-Namespace Communication

Services use Fully Qualified Domain Names (FQDN):

**Format:** `<service-name>.<namespace>.svc.cluster.local:<port>`

**Examples:**
- API Gateway → User Service: `http://user-service.user-service.svc.cluster.local:8081`
- User Service → Database: `postgres-user.infra.svc.cluster.local:5432`
- Payment Service → Kafka: `kafka.infra.svc.cluster.local:9092`

## Troubleshooting

### Check service connectivity
```bash
# From API Gateway pod, test connection to User Service
kubectl exec -it -n api-gateway <api-gateway-pod> -- curl http://user-service.user-service.svc.cluster.local:8081/health
```

### Check logs
```bash
kubectl logs -n user-service <pod-name>
kubectl logs -n infrastructure <postgres-pod-name>
```

### DNS Test
```bash
# Test if services can resolve each other
kubectl run -it --rm debug --image=busybox --restart=Never -n api-gateway -- nslookup user-service.user-service.svc.cluster.local
```

## Clean Up

Delete everything:
```bash
kubectl delete namespace infrastructure
kubectl delete namespace api-gateway
kubectl delete namespace user-service
kubectl delete namespace account-service
kubectl delete namespace payment-service
kubectl delete namespace transfer-service
```
