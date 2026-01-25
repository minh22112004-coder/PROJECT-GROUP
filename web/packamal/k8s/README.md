# Kubernetes Deployment Manifests
## Package Analysis Web - K8s Configuration

This directory contains all Kubernetes manifests for deploying the Package Analysis Web application on Digital Ocean Kubernetes (DOKS).

---

## Directory Structure

```
k8s/
├── README.md                    # This file
├── STRATEGY_PLAN.md            # Comprehensive strategy plan
├── ARCHITECTURE.md             # Architecture design document
├── CAPACITY_PLANNING.md        # Capacity calculations
├── DEPLOYMENT_GUIDE.md         # Deployment instructions
├── namespace.yaml              # Namespace definition
├── configmap.yaml              # Application configuration
├── secrets.yaml.example        # Secrets template
├── api-deployment.yaml         # API server deployment
├── api-service.yaml            # API service
├── api-hpa.yaml                # API auto-scaling
├── worker-deployment.yaml      # Worker deployment
├── worker-hpa.yaml             # Worker auto-scaling
├── postgres-statefulset.yaml   # PostgreSQL database
├── redis-statefulset.yaml      # Redis queue
├── storage-pvc.yaml            # Persistent volumes
├── ingress.yaml                # Ingress configuration
└── network-policy.yaml         # Network security policies
```

---

## Quick Start

### Prerequisites

1. **Kubernetes Cluster**
   - Digital Ocean Kubernetes (DOKS) cluster
   - kubectl configured to access cluster
   - Minimum 3 nodes (4 CPU, 16GB RAM each)

2. **Container Registry**
   - Docker images built and pushed to registry
   - Update image references in deployment files

3. **Secrets**
   - Create secrets from `secrets.yaml.example`
   - Update with actual values

### Deployment Steps

1. **Create Namespace**
   ```bash
   kubectl apply -f namespace.yaml
   ```

2. **Create Secrets**
   ```bash
   # Copy and edit secrets.yaml.example
   cp secrets.yaml.example secrets.yaml
   # Edit secrets.yaml with actual values
   kubectl apply -f secrets.yaml
   ```

3. **Create ConfigMap**
   ```bash
   kubectl apply -f configmap.yaml
   ```

4. **Deploy Storage**
   ```bash
   kubectl apply -f storage-pvc.yaml
   ```

5. **Deploy Database**
   ```bash
   kubectl apply -f postgres-statefulset.yaml
   # Wait for database to be ready
   kubectl wait --for=condition=ready pod -l app=postgres -n package-analysis --timeout=300s
   ```

6. **Deploy Redis**
   ```bash
   kubectl apply -f redis-statefulset.yaml
   # Wait for Redis to be ready
   kubectl wait --for=condition=ready pod -l app=redis -n package-analysis --timeout=300s
   ```

7. **Deploy API**
   ```bash
   kubectl apply -f api-service.yaml
   kubectl apply -f api-deployment.yaml
   kubectl apply -f api-hpa.yaml
   ```

8. **Deploy Workers**
   ```bash
   kubectl apply -f worker-deployment.yaml
   kubectl apply -f worker-hpa.yaml
   ```

9. **Deploy Ingress**
   ```bash
   kubectl apply -f ingress.yaml
   ```

10. **Apply Network Policies**
    ```bash
    kubectl apply -f network-policy.yaml
    ```

---

## Configuration

### Updating Image References

Before deploying, update image references in:
- `api-deployment.yaml`: `your-registry/package-analysis-api:latest`
- `worker-deployment.yaml`: `your-registry/package-analysis-worker:latest`

### Environment Variables

Edit `configmap.yaml` to adjust:
- Queue settings
- Timeout values
- Log levels
- Other application settings

### Resource Limits

Adjust resource requests/limits in:
- `api-deployment.yaml`: API pod resources
- `worker-deployment.yaml`: Worker pod resources
- `postgres-statefulset.yaml`: Database resources
- `redis-statefulset.yaml`: Redis resources

### Auto-scaling

Configure HPA behavior in:
- `api-hpa.yaml`: API scaling (min: 3, max: 10)
- `worker-hpa.yaml`: Worker scaling (min: 0, max: 1000)

---

## Verification

### Check Pod Status
```bash
kubectl get pods -n package-analysis
```

### Check Services
```bash
kubectl get svc -n package-analysis
```

### Check HPA
```bash
kubectl get hpa -n package-analysis
```

### Check Ingress
```bash
kubectl get ingress -n package-analysis
```

### View Logs
```bash
# API logs
kubectl logs -f deployment/package-analysis-api -n package-analysis

# Worker logs
kubectl logs -f deployment/package-analysis-worker -n package-analysis
```

---

## Troubleshooting

### Pods Not Starting
1. Check pod status: `kubectl describe pod <pod-name> -n package-analysis`
2. Check events: `kubectl get events -n package-analysis`
3. Check logs: `kubectl logs <pod-name> -n package-analysis`

### Database Connection Issues
1. Verify PostgreSQL is running: `kubectl get pods -l app=postgres -n package-analysis`
2. Check database credentials in secrets
3. Test connection: `kubectl exec -it <postgres-pod> -n package-analysis -- psql -U postgres`

### Redis Connection Issues
1. Verify Redis is running: `kubectl get pods -l app=redis -n package-analysis`
2. Test connection: `kubectl exec -it <redis-pod> -n package-analysis -- redis-cli ping`

### Auto-scaling Not Working
1. Check HPA status: `kubectl describe hpa <hpa-name> -n package-analysis`
2. Verify metrics server is installed
3. Check pod metrics: `kubectl top pods -n package-analysis`

### Storage Issues
1. Check PVC status: `kubectl get pvc -n package-analysis`
2. Verify storage class exists: `kubectl get storageclass`
3. Check volume mounts: `kubectl describe pod <pod-name> -n package-analysis`

---

## Maintenance

### Rolling Updates
```bash
# Update API image
kubectl set image deployment/package-analysis-api \
  api=your-registry/package-analysis-api:v2.0.0 \
  -n package-analysis

# Update Worker image
kubectl set image deployment/package-analysis-worker \
  worker=your-registry/package-analysis-worker:v2.0.0 \
  -n package-analysis
```

### Scaling Manually
```bash
# Scale API pods
kubectl scale deployment/package-analysis-api --replicas=5 -n package-analysis

# Scale Worker pods
kubectl scale deployment/package-analysis-worker --replicas=10 -n package-analysis
```

### Database Backup
```bash
# Backup PostgreSQL
kubectl exec -it <postgres-pod> -n package-analysis -- \
  pg_dump -U postgres packamal > backup.sql
```

### Redis Backup
```bash
# Create Redis snapshot
kubectl exec -it <redis-pod> -n package-analysis -- redis-cli BGSAVE
```

---

## Security Notes

1. **Secrets Management**
   - Never commit `secrets.yaml` to version control
   - Use external secret management (e.g., Sealed Secrets, Vault)
   - Rotate secrets regularly

2. **Network Policies**
   - Network policies restrict pod-to-pod communication
   - Review and adjust as needed

3. **RBAC**
   - Configure RBAC for cluster access
   - Use least-privilege principle

4. **Image Security**
   - Scan container images for vulnerabilities
   - Use signed images
   - Keep images updated

---

## Cost Optimization

1. **Scale to Zero**
   - Workers scale to 0 during low demand
   - Saves costs during off-peak hours

2. **Right-sizing**
   - Monitor resource usage
   - Adjust requests/limits based on actual usage

3. **Node Pools**
   - Use dedicated node pools for different workloads
   - Consider spot instances for workers

---

## Support

For issues or questions:
1. Review documentation in this directory
2. Check logs and events
3. Consult Digital Ocean Kubernetes documentation
4. Contact DevOps team

---

**Last Updated:** 2025-01-27

