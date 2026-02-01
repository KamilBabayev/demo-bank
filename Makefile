# Makefile for demo-bank

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make help     - Show this help message"
	@echo "  make destroy  - Delete Minikube cluster (all profiles)"
	@echo "  make port-forward - Port-forward api-gateway service to localhost:8080"
	@echo "  make start-minikube - Start Minikube cluster"
	@echo "  make stop-minikube  - Stop Minikube cluster"

.DEFAULT_GOAL := help

.PHONY: destroy
destroy:
	@echo "Deleting Minikube cluster (all profiles)..."
	minikube delete
	@echo "Minikube cluster deleted."	
	
.PHONY: port-forward
port-forward:
	@echo "Port-forwarding api-gateway service to localhost:8080..."
	kubectl port-forward -n api-gateway svc/api-gateway 8080:8080 &
	@echo "Port-forwarding established. Access the API Gateway at http://localhost:8080"

.PHONY: start-minikube
start-minikube:
	@echo "Starting Minikube cluster..."
	minikube start
	@echo "Minikube cluster started."

.PHONY: start-multi-node-minikube
start-multi-node-minikube:
	@echo "Starting multi-node Minikube cluster..."
	minikube start --nodes 3 --cpus 4  --disk-size 10g --driver docker --kubernetes-version v1.34
	@echo "Multi-node Minikube cluster started."

.PHONY: stop-minikube
stop-minikube:
	@echo "Stopping Minikube cluster..."
	minikube stop
	@echo "Minikube cluster stopped."

.PHONY: show-api-gateway-logs
show-api-gateway-logs:
	@echo "Showing logs for api-gateway pod..."
	kubectl logs -n api-gateway -l app=api-gateway --tail=100 -f	


