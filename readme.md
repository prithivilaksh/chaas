# Consistent Hashing as a Service (CHAAS)

A Distributed, highly scalable and stateless service for Consistent Hashing.

## Architecture

- **Master Service**: Manages the hash ring and node distribution
- **Cache Service**: Multiple instances (default 4) that handle cache requests
- **Load Balancing**: Automatic load balancing across cache instances

## Prerequisites

- Go 1.24 or later
- Docker
- Kubernetes (minikube for local development)
- kubectl
- Postman (for testing gRPC endpoints)

## Setup

1. Start minikube:
```bash
minikube start
```

2. Build Docker images:
```bash
docker build -t chaas-master:latest -f Dockerfile.master .
docker build -t chaas-cache:latest -f Dockerfile.cache .
```

3. Load images into minikube:
```bash
minikube image load chaas-master:latest
minikube image load chaas-cache:latest
```

4. Deploy to Kubernetes:
```bash
kubectl apply -f k8s/master-deployment.yaml
kubectl apply -f k8s/cache-deployment.yaml
```

5. Verify deployments:
```bash
kubectl get pods
kubectl get services
```

## Accessing the Services

### Using Port Forwarding

1. Forward the cache service port:
```bash
kubectl port-forward service/cache 50052:50052
```

2. Forward the master service port:
```bash
kubectl port-forward service/master 50051:50051
```

### Service Endpoints

- Master Service: localhost:50051
- Cache Service: localhost:50052

## Testing with Postman

1. Open Postman and create a new gRPC request
2. Enter the service URL (e.g., `localhost:50052` for cache service)
3. Import the proto files:
   - For Cache Service: `cache/grpc/cache.proto`
   - For Master Service: `master/grpc/master.proto`

### Example gRPC Requests

#### Cache Service (localhost:50052)

1. GetState
```json
{
    // No request body needed
}
```

2. GetNextNodeIdByKey
```json
{
    "key": "test-key"
}
```

3. UpdateCache
```json
{
    "hash": "abc123",
    "nodeId": "node1"
}
```

#### Master Service (localhost:50051)

1. CreateHashRing
```json
{
    "numNodes": 4
}
```

2. AddNode
```json
{
    "hash": "abc123"
}
```

3. RemoveNode
```json
{
    "nodeId": "node1"
}
```

4. GetState
```json
{
    // No request body needed
}
```

## Available gRPC Endpoints

### Master Service
- `CreateHashRing`: Create a new hash ring with specified number of nodes
- `AddNode`: Add a new node to the hash ring
- `RemoveNode`: Remove a node from the hash ring
- `GetState`: Get the current state of the hash ring

### Cache Service
- `GetNextNodeIdByKey`: Get the next node ID for a given key
- `UpdateCache`: Update the cache with new node information
- `GetState`: Get the current state of the cache

## Troubleshooting

1. If pods are in CrashLoopBackOff:
```bash
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

2. If services are not accessible:
```bash
kubectl get services
minikube service list
```

3. To check if minikube is running:
```bash
minikube status
```

## Cleanup

To remove all deployments:
```bash
kubectl delete -f k8s/master-deployment.yaml
kubectl delete -f k8s/cache-deployment.yaml
```

To stop minikube:
```bash
minikube stop
```
