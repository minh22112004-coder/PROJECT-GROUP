# Deployment Guide
## Package Analysis Web - Kubernetes on Digital Ocean

**Document Version:** 1.0  
**Date:** 2025-01-27

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Pre-Deployment Setup](#pre-deployment-setup)
3. [Cluster Setup](#cluster-setup)
4. [Application Deployment](#application-deployment)
5. [Post-Deployment Configuration](#post-deployment-configuration)
6. [Verification](#verification)
7. [Monitoring Setup](#monitoring-setup)
8. [Troubleshooting](#troubleshooting)
9. [Maintenance](#maintenance)

---

## Prerequisites

### Required Tools
- `kubectl` (v1.28+)
- `doctl` (Digital Ocean CLI)
- `docker` (for building images)
- `git` (for cloning repository)

### Required Accounts
- Digital Ocean account with billing enabled
- Container registry account (Docker Hub, DO Registry, or other)
- Domain name (for ingress)

### Required Knowledge
- Basic Kubernetes concepts
- Docker containerization
- YAML configuration files
- Linux command line

---

## Pre-Deployment Setup

### 1. Install Required Tools

#### Install kubectl
```bash
# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Verify installation
kubectl version --client
```

#### Install doctl
```bash
# Linux
cd ~
wget https://github.com/digitalocean/doctl/releases/download/v1.104.0/doctl-1.104.0-linux-amd64.tar.gz
tar xf doctl-1.104.0-linux-amd64.tar.gz
sudo mv doctl /usr/local/bin

# Verify installation
doctl version
```

#### Authenticate doctl
```bash
doctl auth init
# Enter your Digital Ocean API token when prompted
```

### 2. Build and Push Container Images

#### Build API Image
```bash
cd /path/to/packamal/web/package-analysis-web

# Build Docker image
docker build -t your-registry/package-analysis-api:latest -f Dockerfile.api .

# Tag for version
docker tag your-registry/package-analysis-api:latest \
  your-registry/package-analysis-api:v1.0.0

# Push to registry
docker push your-registry/package-analysis-api:latest
docker push your-registry/package-analysis-api:v1.0.0
```

#### Build Worker Image
```bash
# Build Docker image
docker build -t your-registry/package-analysis-worker:latest -f Dockerfile.worker .

# Tag for version
docker tag your-registry/package-analysis-worker:latest \
  your-registry/package-analysis-worker:v1.0.0

# Push to registry
docker push your-registry/package-analysis-worker:latest
docker push your-registry/package-analysis-worker:v1.0.0
```

#### Update Image References
```bash
# Update api-deployment.yaml
sed -i 's|your-registry/package-analysis-api:latest|your-registry/package-analysis-api:v1.0.0|g' \
  k8s/api-deployment.yaml

# Update worker-deployment.yaml
sed -i 's|your-registry/package-analysis-worker:latest|your-registry/package-analysis-worker:v1.0.0|g' \
  k8s/worker-deployment.yaml
```

---

## Cluster Setup

### 1. Create Digital Ocean Kubernetes Cluster

#### Using doctl
```bash
# Create cluster
doctl kubernetes cluster create package-analysis-cluster \
  --region nyc1 \
  --node-pool "name=api-pool;size=s-4vcpu-16gb;count=3;tag=api" \
  --node-pool "name=worker-pool;size=s-8vcpu-32gb;count=5;tag=worker" \
  --node-pool "name=db-pool;size=s-4vcpu-16gb;count=3;tag=database" \
  --auto-upgrade=true \
  --maintenance-window "day=saturday,time=02:00"

# Get cluster ID
CLUSTER_ID=$(doctl kubernetes cluster list --format ID --no-header | head -1)

# Get kubeconfig
doctl kubernetes cluster kubeconfig save $CLUSTER_ID

# Verify cluster access
kubectl get nodes
```

#### Using Digital Ocean Web UI
1. Go to Digital Ocean Dashboard
2. Navigate to Kubernetes → Create Cluster
3. Configure:
   - **Name:** package-analysis-cluster
   - **Region:** Choose closest region
   - **Kubernetes Version:** Latest stable
   - **Node Pools:**
     - API Pool: 3 nodes, 4 CPU, 16GB RAM
     - Worker Pool: 5 nodes, 8 CPU, 32GB RAM
     - DB Pool: 3 nodes, 4 CPU, 16GB RAM
4. Click "Create Cluster"
5. Download kubeconfig and configure kubectl

### 2. Install Required Add-ons

#### Install NGINX Ingress Controller
```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/do/deploy.yaml

# Wait for ingress controller to be ready
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=300s
```

#### Install Metrics Server (for HPA)
```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Verify metrics server
kubectl get deployment metrics-server -n kube-system
```

#### Install Cert-Manager (for TLS)
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Wait for cert-manager to be ready
kubectl wait --for=condition=ready pod \
  -l app.kubernetes.io/instance=cert-manager \
  -n cert-manager \
  --timeout=300s
```

#### Create Let's Encrypt ClusterIssuer
```bash
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

---

## Application Deployment

### 1. Create Namespace
```bash
kubectl apply -f k8s/namespace.yaml
```

### 2. Create Secrets

#### Create Secrets from File
```bash
# Copy example file
cp k8s/secrets.yaml.example k8s/secrets.yaml

# Edit secrets.yaml with actual values
nano k8s/secrets.yaml

# Apply secrets
kubectl apply -f k8s/secrets.yaml
```

#### Create Secrets via CLI (Alternative)
```bash
kubectl create secret generic package-analysis-secrets \
  --from-literal=database-url='postgresql://user:password@postgres-service:5432/packamal' \
  --from-literal=redis-url='redis://redis-service:6379/0' \
  --from-literal=secret-key='$(openssl rand -base64 32)' \
  --from-literal=postgres-user='postgres' \
  --from-literal=postgres-password='$(openssl rand -base64 32)' \
  --namespace=package-analysis
```

### 3. Create ConfigMap
```bash
kubectl apply -f k8s/configmap.yaml
```

### 4. Deploy Storage
```bash
kubectl apply -f k8s/storage-pvc.yaml

# Verify PVCs
kubectl get pvc -n package-analysis
```

### 5. Deploy Database
```bash
kubectl apply -f k8s/postgres-statefulset.yaml

# Wait for PostgreSQL to be ready
kubectl wait --for=condition=ready pod \
  -l app=postgres \
  -n package-analysis \
  --timeout=600s

# Verify database
kubectl get pods -l app=postgres -n package-analysis
```

#### Initialize Database (if needed)
```bash
# Get PostgreSQL pod name
POSTGRES_POD=$(kubectl get pod -l app=postgres -n package-analysis -o jsonpath='{.items[0].metadata.name}')

# Run migrations
kubectl exec -it $POSTGRES_POD -n package-analysis -- \
  python manage.py migrate

# Create superuser (optional)
kubectl exec -it $POSTGRES_POD -n package-analysis -- \
  python manage.py createsuperuser
```

### 6. Deploy Redis
```bash
kubectl apply -f k8s/redis-statefulset.yaml

# Wait for Redis to be ready
kubectl wait --for=condition=ready pod \
  -l app=redis \
  -n package-analysis \
  --timeout=300s

# Verify Redis
kubectl get pods -l app=redis -n package-analysis
```

### 7. Deploy API
```bash
kubectl apply -f k8s/api-service.yaml
kubectl apply -f k8s/api-deployment.yaml
kubectl apply -f k8s/api-hpa.yaml

# Wait for API pods to be ready
kubectl wait --for=condition=ready pod \
  -l app=package-analysis-api \
  -n package-analysis \
  --timeout=300s

# Verify API
kubectl get pods -l app=package-analysis-api -n package-analysis
```

### 8. Deploy Workers
```bash
kubectl apply -f k8s/worker-deployment.yaml
kubectl apply -f k8s/worker-hpa.yaml

# Verify workers (may start at 0 replicas)
kubectl get pods -l app=package-analysis-worker -n package-analysis
```

### 9. Deploy Ingress
```bash
# Update ingress.yaml with your domain
sed -i 's/api.packguard.dev/your-domain.com/g' k8s/ingress.yaml

kubectl apply -f k8s/ingress.yaml

# Get ingress IP
kubectl get ingress -n package-analysis
```

### 10. Apply Network Policies
```bash
kubectl apply -f k8s/network-policy.yaml
```

---

## Post-Deployment Configuration

### 1. Configure DNS

#### Get Load Balancer IP
```bash
INGRESS_IP=$(kubectl get ingress package-analysis-ingress -n package-analysis -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
echo "Ingress IP: $INGRESS_IP"
```

#### Update DNS Records
```
Type: A
Name: api (or @ for root domain)
Value: <INGRESS_IP>
TTL: 300
```

### 2. Verify TLS Certificate
```bash
# Check certificate status
kubectl get certificate -n package-analysis

# Wait for certificate to be issued (may take a few minutes)
kubectl wait --for=condition=ready certificate \
  package-analysis-tls \
  -n package-analysis \
  --timeout=600s
```

### 3. Test API Endpoints
```bash
# Test health endpoint
curl https://your-domain.com/health/

# Test API endpoint (with API key)
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://your-domain.com/api/v1/queue/status/
```

---

## Verification

### 1. Check Pod Status
```bash
# All pods
kubectl get pods -n package-analysis

# API pods
kubectl get pods -l app=package-analysis-api -n package-analysis

# Worker pods
kubectl get pods -l app=package-analysis-worker -n package-analysis

# Database pods
kubectl get pods -l app=postgres -n package-analysis

# Redis pods
kubectl get pods -l app=redis -n package-analysis
```

### 2. Check Services
```bash
kubectl get svc -n package-analysis
```

### 3. Check HPA
```bash
kubectl get hpa -n package-analysis

# Describe HPA
kubectl describe hpa package-analysis-api-hpa -n package-analysis
kubectl describe hpa package-analysis-worker-hpa -n package-analysis
```

### 4. Check Ingress
```bash
kubectl get ingress -n package-analysis

# Describe ingress
kubectl describe ingress package-analysis-ingress -n package-analysis
```

### 5. Check Storage
```bash
kubectl get pvc -n package-analysis

# Describe PVC
kubectl describe pvc media-storage-pvc -n package-analysis
```

### 6. Test Application
```bash
# Submit a test task
curl -X POST https://your-domain.com/api/v1/analyze/ \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"purl": "pkg:pypi/requests@2.28.1"}'

# Check queue status
curl https://your-domain.com/api/v1/queue/status/

# Check if workers are scaling
kubectl get pods -l app=package-analysis-worker -n package-analysis
```

---

## Monitoring Setup

### 1. Install Prometheus

#### Using Helm
```bash
# Add Prometheus Helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install Prometheus
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace
```

### 2. Install Grafana
```bash
# Grafana is included in kube-prometheus-stack
# Access Grafana
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80

# Default credentials:
# Username: admin
# Password: prom-operator
```

### 3. Configure Dashboards

#### Import Dashboards
1. Access Grafana at http://localhost:3000
2. Go to Dashboards → Import
3. Import dashboards:
   - Kubernetes Cluster Monitoring
   - PostgreSQL Database
   - Redis Monitoring
   - Custom Application Metrics

### 4. Set Up Alerts

#### Configure Alertmanager
```bash
# Edit Alertmanager config
kubectl edit configmap alertmanager -n monitoring

# Add notification channels (Slack, Email, etc.)
```

---

## Troubleshooting

### Common Issues

#### Pods Not Starting
```bash
# Check pod status
kubectl describe pod <pod-name> -n package-analysis

# Check events
kubectl get events -n package-analysis --sort-by='.lastTimestamp'

# Check logs
kubectl logs <pod-name> -n package-analysis
```

#### Database Connection Issues
```bash
# Verify PostgreSQL is running
kubectl get pods -l app=postgres -n package-analysis

# Test connection
POSTGRES_POD=$(kubectl get pod -l app=postgres -n package-analysis -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it $POSTGRES_POD -n package-analysis -- psql -U postgres -d packamal

# Check database credentials
kubectl get secret package-analysis-secrets -n package-analysis -o jsonpath='{.data.database-url}' | base64 -d
```

#### Redis Connection Issues
```bash
# Verify Redis is running
kubectl get pods -l app=redis -n package-analysis

# Test connection
REDIS_POD=$(kubectl get pod -l app=redis -n package-analysis -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it $REDIS_POD -n package-analysis -- redis-cli ping

# Check queue length
kubectl exec -it $REDIS_POD -n package-analysis -- redis-cli LLEN analysis_tasks
```

#### Auto-scaling Not Working
```bash
# Check HPA status
kubectl describe hpa package-analysis-worker-hpa -n package-analysis

# Verify metrics server
kubectl get deployment metrics-server -n kube-system

# Check pod metrics
kubectl top pods -n package-analysis
```

#### Storage Issues
```bash
# Check PVC status
kubectl get pvc -n package-analysis

# Check storage class
kubectl get storageclass

# Check volume mounts
kubectl describe pod <pod-name> -n package-analysis | grep -A 10 Mounts
```

#### Ingress Issues
```bash
# Check ingress status
kubectl describe ingress package-analysis-ingress -n package-analysis

# Check ingress controller
kubectl get pods -n ingress-nginx

# Check TLS certificate
kubectl get certificate -n package-analysis
```

---

## Maintenance

### Rolling Updates

#### Update API
```bash
# Update image
kubectl set image deployment/package-analysis-api \
  api=your-registry/package-analysis-api:v1.1.0 \
  -n package-analysis

# Monitor rollout
kubectl rollout status deployment/package-analysis-api -n package-analysis

# Rollback if needed
kubectl rollout undo deployment/package-analysis-api -n package-analysis
```

#### Update Workers
```bash
# Update image
kubectl set image deployment/package-analysis-worker \
  worker=your-registry/package-analysis-worker:v1.1.0 \
  -n package-analysis

# Monitor rollout
kubectl rollout status deployment/package-analysis-worker -n package-analysis
```

### Database Maintenance

#### Backup Database
```bash
POSTGRES_POD=$(kubectl get pod -l app=postgres -n package-analysis -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it $POSTGRES_POD -n package-analysis -- \
  pg_dump -U postgres packamal > backup-$(date +%Y%m%d).sql
```

#### Restore Database
```bash
POSTGRES_POD=$(kubectl get pod -l app=postgres -n package-analysis -o jsonpath='{.items[0].metadata.name}')
kubectl exec -i $POSTGRES_POD -n package-analysis -- \
  psql -U postgres packamal < backup-20250127.sql
```

### Scaling

#### Manual Scaling
```bash
# Scale API
kubectl scale deployment/package-analysis-api --replicas=5 -n package-analysis

# Scale Workers
kubectl scale deployment/package-analysis-worker --replicas=10 -n package-analysis
```

#### Cluster Autoscaling
```bash
# Enable cluster autoscaling (if using DOKS)
# Configure via Digital Ocean dashboard or doctl
```

---

## Next Steps

1. **Monitor Performance:** Set up dashboards and alerts
2. **Optimize Resources:** Adjust pod resources based on actual usage
3. **Cost Optimization:** Enable scale-to-zero, use spot instances
4. **Security Hardening:** Review network policies, RBAC, secrets management
5. **Backup Strategy:** Set up automated backups for database and Redis
6. **Documentation:** Document runbooks and procedures

---

## Support

For issues or questions:
- Review troubleshooting section
- Check application logs
- Consult Digital Ocean documentation
- Contact DevOps team

---

**Last Updated:** 2025-01-27

