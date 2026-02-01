# Kubernetes Infrastructure Setup

This folder contains Kubernetes manifests for the infrastructure components.

## Components:

1. **PostgreSQL Databases (4 separate instances)**:
   - `postgres-user` - For User Service (admins)
   - `postgres-account` - For Account Service (customers)
   - `postgres-payment` - For Payment Service
   - `postgres-transfer` - For Transfer Service

2. **Zookeeper** - Required for Kafka

3. **Kafka** - Message broker for async communication

## Deploy to Minikube:

### Step 1: Ensure Minikube is running
```bash
minikube status
# If not running:
minikube start
```

### Step 2: Create namespace and deploy all infrastructure
```bash
# From the k8s-infra directory
kubectl apply -f namespace.yaml
kubectl apply -f postgres-user.yaml
kubectl apply -f postgres-account.yaml
kubectl apply -f postgres-payment.yaml
kubectl apply -f postgres-transfer.yaml
kubectl apply -f zookeeper.yaml
kubectl apply -f kafka.yaml
```

### Step 3: Verify everything is running
```bash
# Check all pods are running in infra namespace
kubectl get pods -n infrastructure

# Check all services in infra namespace
kubectl get services -n infrastructure

# Expected output:
# - 4 PostgreSQL pods
# - 1 Zookeeper pod
# - 1 Kafka pod
```

### Step 4: Wait for all pods to be ready
```bash
kubectl wait --for=condition=ready pod -l app=postgres-user -n infrastructure --timeout=120s
kubectl wait --for=condition=ready pod -l app=postgres-account -n infrastructure --timeout=120s
kubectl wait --for=condition=ready pod -l app=postgres-payment -n infrastructure --timeout=120s
kubectl wait --for=condition=ready pod -l app=postgres-transfer -n infrastructure --timeout=120s
kubectl wait --for=condition=ready pod -l app=zookeeper -n infrastructure --timeout=120s
kubectl wait --for=condition=ready pod -l app=kafka -n infrastructure --timeout=120s
```

## Service DNS Names (FQDN for cross-namespace access):

- User DB: `postgres-user.infra.svc.cluster.local:5432`
- Account DB: `postgres-account.infra.svc.cluster.local:5432`
- Payment DB: `postgres-payment.infra.svc.cluster.local:5432`
- Transfer DB: `postgres-transfer.infra.svc.cluster.local:5432`
- Kafka: `kafka.infra.svc.cluster.local:9092`
- Zookeeper: `zookeeper.infra.svc.cluster.local:2181`

## Clean Up (delete everything):

```bash
kubectl delete -f .
```
