# Capacity Planning & Scaling Calculations
## Package Analysis Web - Kubernetes Deployment

**Document Version:** 1.0  
**Date:** 2025-01-27

---

## Executive Summary

This document provides detailed capacity planning calculations for handling 50,000 requests per day with auto-scaling capabilities on Digital Ocean Kubernetes.

### Key Metrics
- **Daily Requests:** 50,000
- **Monthly Requests:** 1,500,000 (50,000 × 30 days)
- **Average Processing Time:** 10 minutes per task
- **Container Resources:** 2 CPU, 8GB RAM minimum
- **Processing Time Range:** 3-30 minutes per task

---

## Request Distribution Analysis

### Daily Request Pattern

#### Assumptions
- **Total Daily Requests:** 50,000
- **Peak Hours:** 8 hours (33% of day)
- **Off-Peak Hours:** 16 hours (67% of day)
- **Peak Load Factor:** 40% of daily requests in peak hours

#### Distribution
```
Peak Hours (8 hours):
- Requests: 50,000 × 40% = 20,000 requests
- Requests per hour: 20,000 / 8 = 2,500 requests/hour
- Requests per minute: 2,500 / 60 = ~42 requests/minute

Off-Peak Hours (16 hours):
- Requests: 50,000 × 60% = 30,000 requests
- Requests per hour: 30,000 / 16 = 1,875 requests/hour
- Requests per minute: 1,875 / 60 = ~31 requests/minute

Average:
- Requests per hour: 50,000 / 24 = 2,083 requests/hour
- Requests per minute: 2,083 / 60 = ~35 requests/minute
```

### Peak Load Scenario

#### Worst Case (Peak Hour)
```
Peak Hour Requests: 2,500 requests/hour
Average Processing Time: 10 minutes
Concurrent Tasks Needed: 2,500 / 6 = ~417 concurrent tasks

With 150% buffer for spikes:
Concurrent Tasks: 417 × 1.5 = ~625 concurrent workers needed
```

#### Average Load Scenario
```
Average Hour Requests: 2,083 requests/hour
Average Processing Time: 10 minutes
Concurrent Tasks Needed: 2,083 / 6 = ~347 concurrent tasks

With 120% buffer:
Concurrent Tasks: 347 × 1.2 = ~416 concurrent workers needed
```

#### Low Load Scenario (Off-Peak)
```
Off-Peak Hour Requests: 1,875 requests/hour
Average Processing Time: 10 minutes
Concurrent Tasks Needed: 1,875 / 6 = ~313 concurrent tasks

With 110% buffer:
Concurrent Tasks: 313 × 1.1 = ~344 concurrent workers needed
```

---

## Resource Calculations

### Worker Pod Resources

#### Per Worker Pod
- **CPU:** 2 cores (2000m)
- **Memory:** 8GB (8Gi)
- **Storage:** Minimal (ephemeral)

#### Total Worker Resources

##### Peak Load (625 workers)
```
CPU: 625 × 2 = 1,250 CPU cores
Memory: 625 × 8 = 5,000 GB (5 TB)
```

##### Average Load (416 workers)
```
CPU: 416 × 2 = 832 CPU cores
Memory: 416 × 8 = 3,328 GB (3.3 TB)
```

##### Low Load (344 workers)
```
CPU: 344 × 2 = 688 CPU cores
Memory: 344 × 8 = 2,752 GB (2.7 TB)
```

### API Pod Resources

#### Per API Pod
- **CPU:** 1 core (1000m)
- **Memory:** 2GB (2Gi)

#### API Load Calculation
```
Peak Requests: 2,500 requests/hour = 42 requests/minute
API Capacity per Pod: 100 requests/minute (with 70% CPU target)
Pods Needed: 42 / 100 = ~1 pod (with buffer: 3-5 pods)

Average Requests: 2,083 requests/hour = 35 requests/minute
Pods Needed: 35 / 100 = ~1 pod (with buffer: 3 pods)
```

#### Total API Resources
```
Peak: 5 pods × 1 CPU = 5 CPU cores, 5 × 2GB = 10GB
Average: 3 pods × 1 CPU = 3 CPU cores, 3 × 2GB = 6GB
```

### Database Resources

#### PostgreSQL Primary
- **CPU:** 4 cores
- **Memory:** 16GB
- **Storage:** 500GB SSD (auto-scaling)

#### PostgreSQL Replicas (2)
- **CPU:** 2 cores each (4 total)
- **Memory:** 8GB each (16GB total)
- **Storage:** 100GB each (200GB total)

#### Total Database Resources
```
CPU: 4 + 4 = 8 CPU cores
Memory: 16 + 16 = 32GB
Storage: 500 + 200 = 700GB
```

### Redis Resources

#### Redis Master
- **CPU:** 2 cores
- **Memory:** 4GB
- **Storage:** 50GB SSD

#### Redis Replicas (2)
- **CPU:** 2 cores each (4 total)
- **Memory:** 4GB each (8GB total)
- **Storage:** 50GB each (100GB total)

#### Total Redis Resources
```
CPU: 2 + 4 = 6 CPU cores
Memory: 4 + 8 = 12GB
Storage: 50 + 100 = 150GB
```

---

## Node Capacity Planning

### Node Specifications

#### API Node Pool
- **Node Size:** 4 CPU, 16GB RAM
- **Usable Resources:** ~3.5 CPU, 14GB RAM (accounting for system overhead)
- **Pods per Node:** 3-4 API pods

#### Worker Node Pool
- **Node Size:** 8 CPU, 32GB RAM
- **Usable Resources:** ~7 CPU, 28GB RAM
- **Pods per Node:** 3 worker pods (2 CPU, 8GB each)

#### Database Node Pool
- **Node Size:** 4 CPU, 16GB RAM
- **Usable Resources:** ~3.5 CPU, 14GB RAM
- **Pods per Node:** 1 database pod

### Node Count Calculations

#### API Nodes
```
Peak: 5 pods / 3 pods per node = ~2 nodes (with buffer: 3-5 nodes)
Average: 3 pods / 3 pods per node = 1 node (with buffer: 3 nodes)
Min Nodes: 3 (for HA)
Max Nodes: 10 (HPA limit)
```

#### Worker Nodes
```
Peak: 625 pods / 3 pods per node = ~208 nodes
Average: 416 pods / 3 pods per node = ~139 nodes
Low: 344 pods / 3 pods per node = ~115 nodes

Min Nodes: 0 (scale to zero)
Max Nodes: 333 (1000 pods / 3 pods per node)
```

#### Database Nodes
```
Fixed: 3 nodes (1 primary + 2 replicas)
```

### Total Node Count

#### Peak Load
```
API Nodes: 5
Worker Nodes: 208
Database Nodes: 3
Total: 216 nodes
```

#### Average Load
```
API Nodes: 3
Worker Nodes: 139
Database Nodes: 3
Total: 145 nodes
```

#### Low Load
```
API Nodes: 3
Worker Nodes: 115
Database Nodes: 3
Total: 121 nodes
```

---

## Storage Calculations

### Database Storage

#### Current Data
- **Task Records:** ~1,500,000 tasks/month
- **Record Size:** ~2KB per task
- **Monthly Growth:** 1,500,000 × 2KB = 3GB/month
- **Annual Growth:** 3GB × 12 = 36GB/year

#### Storage Requirements
```
Initial: 100GB
Monthly Growth: 3GB
6 Months: 100 + (3 × 6) = 118GB
1 Year: 100 + 36 = 136GB
3 Years: 100 + (36 × 3) = 208GB

Recommended: 500GB (with 2.4x growth buffer)
```

### Media Storage

#### Report Files
- **Reports per Task:** 1 JSON file
- **Average Report Size:** 50KB
- **Monthly Reports:** 1,500,000
- **Monthly Growth:** 1,500,000 × 50KB = 75GB/month
- **Annual Growth:** 75GB × 12 = 900GB/year

#### Storage Requirements
```
Initial: 100GB
Monthly Growth: 75GB
6 Months: 100 + (75 × 6) = 550GB
1 Year: 100 + 900 = 1TB
3 Years: 100 + (900 × 3) = 2.8TB

Recommended: 1TB+ (with auto-scaling)
```

### Redis Storage

#### Queue Data
- **Tasks in Queue:** Max 1,000 (typical)
- **Task Data Size:** ~1KB per task
- **Queue Storage:** 1,000 × 1KB = 1MB
- **Overhead:** 10x for Redis overhead = 10MB

#### Storage Requirements
```
Queue Data: 10MB
AOF (Append Only File): 50MB
Snapshots: 100MB
Buffer: 50GB (for growth and performance)

Recommended: 50GB
```

---

## Cost Calculations

### Digital Ocean Pricing (Monthly)

#### Node Costs

##### API Nodes
```
Node Size: 4 CPU, 16GB RAM = $48/month per node
Peak: 5 nodes × $48 = $240/month
Average: 3 nodes × $48 = $144/month
```

##### Worker Nodes
```
Node Size: 8 CPU, 32GB RAM = $96/month per node
Peak: 208 nodes × $96 = $19,968/month
Average: 139 nodes × $96 = $13,344/month
Low: 115 nodes × $96 = $11,040/month
```

##### Database Nodes
```
Node Size: 4 CPU, 16GB RAM = $48/month per node
Fixed: 3 nodes × $48 = $144/month
```

##### Total Node Costs
```
Peak: $240 + $19,968 + $144 = $20,352/month
Average: $144 + $13,344 + $144 = $13,632/month
Low: $144 + $11,040 + $144 = $11,328/month
```

#### Storage Costs

##### Block Storage (SSD)
```
PostgreSQL: 500GB × $0.10/GB = $50/month
Redis: 50GB × $0.10/GB = $5/month
Total: $55/month
```

##### Object Storage (DO Spaces)
```
Media Files: 1TB × $5/TB = $5/month
Additional: $0.02/GB for transfer = ~$10/month
Total: $15/month
```

##### Total Storage Costs
```
$55 + $15 = $70/month
```

#### Load Balancer
```
Digital Ocean Load Balancer: $12/month
```

#### Total Monthly Costs

##### Peak Load
```
Nodes: $20,352
Storage: $70
Load Balancer: $12
Total: $20,434/month
```

##### Average Load
```
Nodes: $13,632
Storage: $70
Load Balancer: $12
Total: $13,714/month
```

##### Low Load
```
Nodes: $11,328
Storage: $70
Load Balancer: $12
Total: $11,410/month
```

### Cost Per Request

#### Average Monthly Cost
```
Average Load Cost: $13,714/month
Monthly Requests: 1,500,000
Cost per Request: $13,714 / 1,500,000 = $0.0091/request
```

#### Cost Optimization with Auto-scaling
```
If workers scale to zero during 8 hours of low demand:
- Savings: 8 hours / 24 hours = 33% reduction
- Optimized Cost: $13,714 × 0.67 = $9,188/month
- Cost per Request: $9,188 / 1,500,000 = $0.0061/request
```

---

## Auto-scaling Strategy

### Queue-Based Scaling

#### Scaling Formula
```
if queue_length < 10:
    desired_replicas = 1
elif queue_length < 50:
    desired_replicas = queue_length / 10
elif queue_length < 200:
    desired_replicas = queue_length / 5
elif queue_length < 500:
    desired_replicas = queue_length / 3
else:
    desired_replicas = queue_length / 2
```

#### Scaling Examples

##### Low Queue (10 tasks)
```
desired_replicas = 1
Actual: 1 worker pod
```

##### Medium Queue (100 tasks)
```
desired_replicas = 100 / 5 = 20
Actual: 20 worker pods
```

##### High Queue (500 tasks)
```
desired_replicas = 500 / 3 = ~167
Actual: 167 worker pods
```

##### Very High Queue (1000 tasks)
```
desired_replicas = 1000 / 2 = 500
Actual: 500 worker pods
```

### Time-Based Scaling

#### Scale Down During Off-Peak
```
Off-Peak Hours: 16 hours/day
Scale Down Factor: 50%
Workers: 416 × 0.5 = 208 workers
Cost Savings: 50% during off-peak
```

#### Scale to Zero
```
If queue_length = 0 for 30 minutes:
    desired_replicas = 0
    Cost Savings: 100% (no worker nodes)
```

---

## Performance Targets

### Response Time Targets
- **API Response Time (p95):** < 500ms
- **API Response Time (p99):** < 1s
- **Task Processing Time:** 3-30 minutes (as designed)
- **Queue Wait Time (p95):** < 5 minutes

### Throughput Targets
- **API Requests/Second:** 100 req/s per pod
- **Tasks Processed/Hour:** 6 tasks per worker (10 min average)
- **Peak Capacity:** 625 workers × 6 = 3,750 tasks/hour

### Availability Targets
- **Uptime:** 99.9% (8.76 hours downtime/year)
- **Database Uptime:** 99.95% (4.38 hours downtime/year)
- **Queue Uptime:** 99.95% (4.38 hours downtime/year)

---

## Capacity Planning Summary

### Resource Summary Table

| Component | Peak | Average | Low | Min | Max |
|-----------|------|---------|-----|-----|-----|
| **API Pods** | 5 | 3 | 3 | 3 | 10 |
| **Worker Pods** | 625 | 416 | 344 | 0 | 1000 |
| **API Nodes** | 5 | 3 | 3 | 3 | 10 |
| **Worker Nodes** | 208 | 139 | 115 | 0 | 333 |
| **Database Nodes** | 3 | 3 | 3 | 3 | 3 |
| **Total Nodes** | 216 | 145 | 121 | 6 | 346 |

### Storage Summary Table

| Storage Type | Size | Growth Rate | Recommended |
|--------------|------|-------------|-------------|
| **PostgreSQL** | 500GB | 3GB/month | 500GB (auto-scale) |
| **Redis** | 50GB | Minimal | 50GB |
| **Media Files** | 1TB | 75GB/month | 1TB+ (auto-scale) |

### Cost Summary Table

| Scenario | Monthly Cost | Cost/Request | Notes |
|----------|--------------|--------------|-------|
| **Peak Load** | $20,434 | $0.0136 | Maximum capacity |
| **Average Load** | $13,714 | $0.0091 | Typical usage |
| **Low Load** | $11,410 | $0.0076 | Off-peak hours |
| **Optimized** | $9,188 | $0.0061 | With scale-to-zero |

---

## Recommendations

### Immediate Actions
1. **Start with Average Load:** Deploy with 3 API nodes, 139 worker nodes
2. **Enable Auto-scaling:** Configure HPA for automatic scaling
3. **Monitor Metrics:** Set up monitoring for queue length, CPU, memory
4. **Set Budget Alerts:** Configure alerts at $15,000/month threshold

### Optimization Opportunities
1. **Scale to Zero:** Enable scale-to-zero for workers during low demand
2. **Spot Instances:** Use spot/preemptible nodes for workers (50% savings)
3. **Reserved Instances:** Reserve API and DB nodes (20% savings)
4. **Right-sizing:** Monitor and adjust pod resources based on actual usage

### Future Scaling
1. **Multi-Region:** Deploy to multiple regions for global coverage
2. **CDN:** Use CDN for media file delivery
3. **Caching:** Implement Redis caching for frequently accessed data
4. **Database Sharding:** Shard database if growth exceeds capacity

---

## Appendix

### A. Calculation Formulas

#### Concurrent Tasks
```
concurrent_tasks = (requests_per_hour / 60) × average_processing_time_minutes
```

#### Workers Needed
```
workers_needed = concurrent_tasks × buffer_factor
```

#### Nodes Needed
```
nodes_needed = workers_needed / pods_per_node
```

#### Cost per Request
```
cost_per_request = monthly_cost / monthly_requests
```

### B. Assumptions

1. **Processing Time:** Average 10 minutes, range 3-30 minutes
2. **Peak Hours:** 8 hours (33% of day)
3. **Peak Load Factor:** 40% of daily requests in peak hours
4. **Buffer Factor:** 120-150% for capacity planning
5. **System Overhead:** 10-15% for node resources
6. **Storage Growth:** Linear growth based on historical data

### C. References

- Digital Ocean Pricing: https://www.digitalocean.com/pricing
- Kubernetes HPA Documentation
- PostgreSQL Capacity Planning Guide
- Redis Memory Planning Guide

---

**Document Status:** Final  
**Last Updated:** 2025-01-27

