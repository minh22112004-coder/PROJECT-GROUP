# Kubernetes Architecture Design
## Package Analysis Web - Digital Ocean K8s Deployment

**Document Version:** 1.0  
**Date:** 2025-01-27

---

## Architecture Overview

This document describes the detailed Kubernetes architecture for deploying the Package Analysis Web application on Digital Ocean Kubernetes (DOKS) with horizontal auto-scaling capabilities.

---

## System Architecture Diagram

```mermaid
graph TB
    subgraph "External Layer"
        INTERNET[Internet]
        LB[Digital Ocean Load Balancer<br/>Ingress Controller]
    end

    subgraph "Kubernetes Cluster - DOKS"
        subgraph "Ingress Layer"
            INGRESS[NGINX Ingress Controller]
        end

        subgraph "API Layer - Stateless"
            API1[API Pod 1<br/>Django/Gunicorn<br/>1 CPU, 2GB RAM]
            API2[API Pod 2<br/>Django/Gunicorn<br/>1 CPU, 2GB RAM]
            API3[API Pod N<br/>Django/Gunicorn<br/>1 CPU, 2GB RAM]
            API_SVC[API Service<br/>ClusterIP]
            API_HPA[HPA<br/>Min: 3, Max: 10]
        end

        subgraph "Worker Layer - Stateless"
            W1[Worker Pod 1<br/>Task Processor<br/>2 CPU, 8GB RAM]
            W2[Worker Pod 2<br/>Task Processor<br/>2 CPU, 8GB RAM]
            WN[Worker Pod N<br/>Task Processor<br/>2 CPU, 8GB RAM]
            WORKER_HPA[HPA<br/>Min: 0, Max: 1000<br/>Queue-based Scaling]
        end

        subgraph "Data Layer - Stateful"
            POSTGRES_PRIMARY[(PostgreSQL Primary<br/>StatefulSet<br/>4 CPU, 16GB RAM)]
            POSTGRES_REPLICA1[(PostgreSQL Replica 1<br/>StatefulSet<br/>2 CPU, 8GB RAM)]
            POSTGRES_REPLICA2[(PostgreSQL Replica 2<br/>StatefulSet<br/>2 CPU, 8GB RAM)]
            POSTGRES_SVC[PostgreSQL Service<br/>ClusterIP]
        end

        subgraph "Queue Layer - Stateful"
            REDIS_MASTER[(Redis Master<br/>StatefulSet<br/>2 CPU, 4GB RAM)]
            REDIS_REPLICA1[(Redis Replica 1<br/>StatefulSet<br/>2 CPU, 4GB RAM)]
            REDIS_REPLICA2[(Redis Replica 2<br/>StatefulSet<br/>2 CPU, 4GB RAM)]
            REDIS_SVC[Redis Service<br/>ClusterIP]
        end

        subgraph "Storage Layer"
            PV1[Persistent Volume 1<br/>PostgreSQL Data<br/>500GB SSD]
            PV2[Persistent Volume 2<br/>Redis Data<br/>50GB SSD]
            PV3[Persistent Volume 3<br/>Media Files<br/>NFS/DO Spaces]
        end

        subgraph "Monitoring & Observability"
            PROMETHEUS[Prometheus<br/>Metrics Collection]
            GRAFANA[Grafana<br/>Visualization]
            LOKI[Loki<br/>Log Aggregation]
            ALERTMANAGER[Alertmanager<br/>Alerting]
        end
    end

    subgraph "External Services"
        REGISTRY[Container Registry<br/>Docker Hub / DO Registry]
        DO_SPACES[Digital Ocean Spaces<br/>S3-compatible Storage]
        PACKAGE_REPOS[Package Repositories<br/>PyPI, npm, etc.]
    end

    INTERNET --> LB
    LB --> INGRESS
    INGRESS --> API_SVC
    API_SVC --> API1
    API_SVC --> API2
    API_SVC --> API3
    API_HPA -.-> API1
    API_HPA -.-> API2
    API_HPA -.-> API3

    API1 --> REDIS_SVC
    API2 --> REDIS_SVC
    API3 --> REDIS_SVC
    REDIS_SVC --> REDIS_MASTER
    REDIS_MASTER --> REDIS_REPLICA1
    REDIS_MASTER --> REDIS_REPLICA2

    API1 --> POSTGRES_SVC
    API2 --> POSTGRES_SVC
    API3 --> POSTGRES_SVC
    POSTGRES_SVC --> POSTGRES_PRIMARY
    POSTGRES_PRIMARY --> POSTGRES_REPLICA1
    POSTGRES_PRIMARY --> POSTGRES_REPLICA2

    W1 --> REDIS_SVC
    W2 --> REDIS_SVC
    WN --> REDIS_SVC
    W1 --> POSTGRES_SVC
    W2 --> POSTGRES_SVC
    WN --> POSTGRES_SVC
    WORKER_HPA -.-> W1
    WORKER_HPA -.-> W2
    WORKER_HPA -.-> WN

    POSTGRES_PRIMARY --> PV1
    REDIS_MASTER --> PV2
    API1 --> PV3
    API2 --> PV3
    API3 --> PV3
    W1 --> PV3
    W2 --> PV3
    WN --> PV3

    W1 --> REGISTRY
    W2 --> REGISTRY
    WN --> REGISTRY
    W1 --> PACKAGE_REPOS
    W2 --> PACKAGE_REPOS
    WN --> PACKAGE_REPOS

    API1 --> DO_SPACES
    API2 --> DO_SPACES
    API3 --> DO_SPACES
    W1 --> DO_SPACES
    W2 --> DO_SPACES
    WN --> DO_SPACES

    PROMETHEUS --> API1
    PROMETHEUS --> API2
    PROMETHEUS --> API3
    PROMETHEUS --> W1
    PROMETHEUS --> W2
    PROMETHEUS --> WN
    PROMETHEUS --> POSTGRES_PRIMARY
    PROMETHEUS --> REDIS_MASTER
    GRAFANA --> PROMETHEUS
    LOKI --> API1
    LOKI --> API2
    LOKI --> API3
    LOKI --> W1
    LOKI --> W2
    LOKI --> WN
    ALERTMANAGER --> PROMETHEUS

    style API1 fill:#e1f5ff
    style API2 fill:#e1f5ff
    style API3 fill:#e1f5ff
    style W1 fill:#ffe1f5
    style W2 fill:#ffe1f5
    style WN fill:#ffe1f5
    style POSTGRES_PRIMARY fill:#e1ffe1
    style REDIS_MASTER fill:#fff4e1
    style PROMETHEUS fill:#f0e1ff
    style GRAFANA fill:#f0e1ff
```

---

## Component Details

### 1. API Layer (Django/Gunicorn)

#### Deployment Configuration
- **Type:** Stateless Deployment
- **Replicas:** 3-10 (auto-scaling)
- **Image:** `your-registry/package-analysis-api:latest`
- **Resources:**
  - Requests: 0.5 CPU, 1GB RAM
  - Limits: 1 CPU, 2GB RAM
- **Health Checks:**
  - Liveness: `/health/` endpoint
  - Readiness: `/ready/` endpoint
- **Environment Variables:**
  - `DATABASE_URL`: PostgreSQL connection string
  - `REDIS_URL`: Redis connection string
  - `MEDIA_ROOT`: `/media` (mounted volume)
  - `ALLOWED_HOSTS`: Domain names
  - `SECRET_KEY`: Django secret key (from secret)

#### Service Configuration
- **Type:** ClusterIP
- **Port:** 8000
- **Session Affinity:** None (stateless)

#### Auto-scaling (HPA)
```yaml
Min Replicas: 3
Max Replicas: 10
Metrics:
  - CPU: 70% target
  - Memory: 80% target
  - Custom: requests_per_second (100 req/s per pod)
Scale Up: 2 pods per minute
Scale Down: 1 pod per 5 minutes
```

---

### 2. Worker Layer (Task Processors)

#### Deployment Configuration
- **Type:** Stateless Deployment
- **Replicas:** 0-1000 (auto-scaling)
- **Image:** `your-registry/package-analysis-worker:latest`
- **Resources:**
  - Requests: 2 CPU, 8GB RAM
  - Limits: 2 CPU, 8GB RAM
- **Privileged:** Yes (for Docker-in-Docker)
- **Volume Mounts:**
  - `/var/run/docker.sock`: Docker socket
  - `/media`: Shared media storage
- **Environment Variables:**
  - `DATABASE_URL`: PostgreSQL connection string
  - `REDIS_URL`: Redis connection string
  - `QUEUE_NAME`: `analysis_tasks`
  - `WORKER_CONCURRENCY`: 1 (one task per worker)

#### Service Configuration
- **Type:** None (no external access needed)

#### Auto-scaling (HPA)
```yaml
Min Replicas: 0 (scale to zero)
Max Replicas: 1000
Metrics:
  - Custom: queue_length (primary)
  - CPU: 80% target
  - Memory: 85% target
Scale Up: Aggressive (10 pods per minute)
Scale Down: Conservative (1 pod per 10 minutes)
```

#### Queue-Based Scaling Logic
```
if queue_length < 10:
    desired_replicas = 1
elif queue_length < 50:
    desired_replicas = queue_length / 10
elif queue_length < 200:
    desired_replicas = queue_length / 5
else:
    desired_replicas = queue_length / 2
```

---

### 3. Database Layer (PostgreSQL)

#### StatefulSet Configuration
- **Type:** StatefulSet (or Managed PostgreSQL)
- **Replicas:** 1 Primary + 2 Read Replicas
- **Image:** `postgres:15-alpine`
- **Resources:**
  - Primary: 4 CPU, 16GB RAM
  - Replicas: 2 CPU, 8GB RAM
- **Storage:**
  - Primary: 500GB SSD (auto-scaling)
  - Replicas: 100GB SSD each
- **Backup:** Daily automated backups, 30-day retention

#### Service Configuration
- **Type:** ClusterIP
- **Port:** 5432
- **Read Replicas:** Separate service for read operations

#### High Availability
- **Primary-Replica:** Streaming replication
- **Failover:** Automated failover (if using managed service)
- **Connection Pooling:** PgBouncer sidecar (optional)

---

### 4. Queue Layer (Redis)

#### StatefulSet Configuration
- **Type:** StatefulSet (or Managed Redis)
- **Replicas:** 1 Master + 2 Replicas (HA)
- **Image:** `redis:7-alpine`
- **Resources:**
  - Master: 2 CPU, 4GB RAM
  - Replicas: 2 CPU, 4GB RAM
- **Storage:** 50GB SSD (persistent)
- **Persistence:** AOF (Append Only File) enabled

#### Service Configuration
- **Type:** ClusterIP
- **Port:** 6379
- **Sentinel:** Redis Sentinel for HA (optional)

#### Queue Structure
```python
# Queue names
QUEUE_NAME = "analysis_tasks"
PRIORITY_QUEUE = "analysis_tasks:priority"
DELAYED_QUEUE = "analysis_tasks:delayed"

# Task structure
{
    "task_id": 123,
    "purl": "pkg:pypi/requests@2.28.1",
    "priority": 0,
    "created_at": "2025-01-27T10:00:00Z",
    "timeout_minutes": 30
}
```

---

### 5. Storage Layer

#### Persistent Volumes
- **PostgreSQL Data:** 500GB SSD, ReadWriteOnce
- **Redis Data:** 50GB SSD, ReadWriteOnce
- **Media Files:** NFS or DO Spaces, ReadWriteMany

#### Digital Ocean Spaces Integration
- **Purpose:** Media files, analysis reports
- **Access:** S3-compatible API
- **Mount:** Via s3fs or application-level integration
- **Backup:** Automatic replication

#### Volume Claims
```yaml
# PostgreSQL PVC
storageClassName: do-block-storage
accessModes: [ReadWriteOnce]
resources:
  requests:
    storage: 500Gi

# Redis PVC
storageClassName: do-block-storage
accessModes: [ReadWriteOnce]
resources:
  requests:
    storage: 50Gi

# Media PVC (NFS)
storageClassName: do-nfs-storage
accessModes: [ReadWriteMany]
resources:
  requests:
    storage: 1Ti
```

---

### 6. Ingress & Load Balancing

#### Ingress Configuration
- **Controller:** NGINX Ingress Controller
- **Type:** LoadBalancer (Digital Ocean)
- **TLS:** Let's Encrypt certificates
- **Rate Limiting:** 100 requests/second per IP
- **Annotations:**
  - `cert-manager.io/cluster-issuer`: Let's Encrypt
  - `nginx.ingress.kubernetes.io/rate-limit`: 100

#### Routing Rules
```yaml
Host: api.packguard.dev
Paths:
  - /api/* -> API Service (port 8000)
  - /media/* -> Static files (NFS/Spaces)
  - /admin/* -> API Service (port 8000)
  - /health -> API Service (port 8000)
```

---

### 7. Monitoring & Observability

#### Prometheus
- **Purpose:** Metrics collection
- **Scrape Interval:** 15 seconds
- **Retention:** 30 days
- **Targets:**
  - API pods
  - Worker pods
  - PostgreSQL
  - Redis
  - Node metrics

#### Grafana
- **Purpose:** Visualization and dashboards
- **Dashboards:**
  - API performance
  - Worker queue metrics
  - Database performance
  - Infrastructure metrics
  - Cost tracking

#### Loki
- **Purpose:** Log aggregation
- **Retention:** 30 days
- **Sources:**
  - API logs
  - Worker logs
  - System logs

#### Alertmanager
- **Purpose:** Alert routing and notification
- **Channels:**
  - Email
  - Slack
  - PagerDuty

---

## Network Architecture

### Network Policies

#### API Pods
```yaml
Ingress:
  - From: Ingress Controller
  - Port: 8000
Egress:
  - To: PostgreSQL (port 5432)
  - To: Redis (port 6379)
  - To: Internet (HTTPS)
```

#### Worker Pods
```yaml
Ingress: None
Egress:
  - To: PostgreSQL (port 5432)
  - To: Redis (port 6379)
  - To: Container Registry (HTTPS)
  - To: Package Repositories (HTTPS)
```

#### Database Pods
```yaml
Ingress:
  - From: API Pods (port 5432)
  - From: Worker Pods (port 5432)
Egress: None
```

---

## Security Architecture

### Authentication & Authorization
- **API:** Existing API key system (unchanged)
- **Kubernetes:** RBAC for cluster access
- **Database:** Network policies + credentials

### Secrets Management
- **Type:** Kubernetes Secrets (or external vault)
- **Secrets:**
  - Database credentials
  - Redis credentials
  - Django SECRET_KEY
  - API keys
  - Container registry credentials

### Network Security
- **TLS:** All external communication encrypted
- **Internal:** Cluster networking (encrypted)
- **Policies:** Network policies restrict pod communication

### Container Security
- **Image Scanning:** Automated vulnerability scanning
- **Pod Security:** Pod security policies enforced
- **Resource Limits:** Prevent resource exhaustion

---

## Data Flow

### Request Flow
```
1. Client → Load Balancer → Ingress → API Pod
2. API Pod → Check Cache (Redis)
3. If not cached:
   a. Create task in database
   b. Add task to queue (Redis)
   c. Return task_id to client
4. Worker picks task from queue
5. Worker processes task:
   a. Pull Docker image
   b. Run analysis container
   c. Save results to storage
   d. Update task status in database
6. Client polls task status
7. Client downloads result
```

### Queue Processing Flow
```
1. Worker polls Redis queue
2. Worker claims task (atomic operation)
3. Worker updates task status: queued → running
4. Worker starts Docker container
5. Container runs analysis
6. Container saves results
7. Worker updates task status: running → completed
8. Worker removes task from queue
```

---

## Scaling Behavior

### Scale Up Scenario
```
1. Queue length increases to 100
2. HPA detects queue_length metric
3. HPA calculates: 100 / 10 = 10 workers needed
4. HPA scales workers from 5 to 10
5. New pods start and connect to queue
6. Queue length decreases
7. System stabilizes
```

### Scale Down Scenario
```
1. Queue length decreases to 5
2. HPA detects low queue length
3. HPA waits 10 minutes (cooldown)
4. HPA scales workers from 10 to 1
5. Remaining workers handle queue
6. If queue empty, scale to 0
```

### Peak Handling
```
Peak Scenario:
- 20,000 requests in 8 hours
- Average: 2,500 requests/hour
- Peak: 5,000 requests/hour
- Concurrent tasks: ~500

Scaling Response:
- Workers scale to 500 pods
- API pods scale to 10
- Queue processes at max capacity
- System handles peak load
```

---

## Disaster Recovery

### Backup Strategy
- **Database:** Daily automated backups, point-in-time recovery
- **Redis:** Periodic snapshots
- **Media Files:** Replicated to multiple regions
- **Configuration:** Git version control

### Recovery Procedures
- **RTO:** 1 hour
- **RPO:** 24 hours
- **Failover:** Automated for managed services
- **Testing:** Quarterly DR drills

---

## Cost Optimization

### Resource Optimization
- **Right-sizing:** Monitor and adjust pod resources
- **Scale to Zero:** Workers scale to 0 during low demand
- **Reserved Instances:** Reserve API and DB nodes
- **Spot Instances:** Use for worker nodes (optional)

### Monitoring
- **Cost Tracking:** Per-namespace resource usage
- **Alerts:** Budget threshold alerts
- **Reports:** Monthly cost analysis

---

## Deployment Strategy

### Blue-Green Deployment
1. Deploy new version to "green" environment
2. Test green environment
3. Switch traffic from "blue" to "green"
4. Monitor green environment
5. Decommission blue environment

### Rolling Update
1. Update deployment with new image
2. Kubernetes gradually replaces old pods
3. Health checks ensure new pods are ready
4. Old pods terminated after new pods healthy

### Canary Deployment
1. Deploy new version to 10% of pods
2. Monitor metrics and errors
3. Gradually increase to 100%
4. Rollback if issues detected

---

## Appendix

### A. Resource Specifications Summary

| Component | CPU | Memory | Storage | Replicas |
|-----------|-----|--------|---------|----------|
| API Pod | 1 | 2GB | - | 3-10 |
| Worker Pod | 2 | 8GB | - | 0-1000 |
| PostgreSQL Primary | 4 | 16GB | 500GB | 1 |
| PostgreSQL Replica | 2 | 8GB | 100GB | 2 |
| Redis Master | 2 | 4GB | 50GB | 1 |
| Redis Replica | 2 | 4GB | 50GB | 2 |

### B. Port Mappings

| Service | Port | Protocol | Access |
|---------|------|----------|--------|
| API | 8000 | HTTP | External (via Ingress) |
| PostgreSQL | 5432 | TCP | Internal only |
| Redis | 6379 | TCP | Internal only |
| Prometheus | 9090 | HTTP | Internal |
| Grafana | 3000 | HTTP | Internal (or external) |

### C. Environment Variables Reference

#### API Pods
```bash
DATABASE_URL=postgresql://user:pass@postgres:5432/dbname
REDIS_URL=redis://redis:6379/0
MEDIA_ROOT=/media
ALLOWED_HOSTS=api.packguard.dev
SECRET_KEY=<from-secret>
DEBUG=False
```

#### Worker Pods
```bash
DATABASE_URL=postgresql://user:pass@postgres:5432/dbname
REDIS_URL=redis://redis:6379/0
QUEUE_NAME=analysis_tasks
WORKER_CONCURRENCY=1
DOCKER_HOST=unix:///var/run/docker.sock
```

---

**Document Status:** Final  
**Last Updated:** 2025-01-27

