# Capacity Planning

## Overview

This document outlines the capacity planning methodology for the Task Manager API service, including current resource utilization, growth projections, and scaling recommendations.

## Current System Specifications

| Resource       | Specification                      | Current Usage |
| -------------- | ---------------------------------- | ------------- |
| **GKE Nodes**  | e2-standard-2 (2 vCPU, 8GB RAM)    | 2 nodes       |
| **API Pods**   | 100m-500m CPU, 128-256Mi RAM       | 3 replicas    |
| **PostgreSQL** | db-custom-2-4096 (2 vCPU, 4GB RAM) | 20GB storage  |
| **Redis**      | 1GB Memorystore                    | ~200MB used   |
| **MongoDB**    | M10 (2GB RAM)                      | ~5GB storage  |

## Key Metrics & Thresholds

| Metric         | Current   | Warning   | Critical  | Max Capacity |
| -------------- | --------- | --------- | --------- | ------------ |
| Request Rate   | 100 req/s | 500 req/s | 800 req/s | ~1000 req/s  |
| P95 Latency    | 50ms      | 200ms     | 500ms     | —            |
| CPU Usage      | 30%       | 70%       | 85%       | 100%         |
| Memory Usage   | 45%       | 75%       | 90%       | 100%         |
| DB Connections | 20        | 80        | 95        | 100          |
| Redis Memory   | 200MB     | 800MB     | 950MB     | 1GB          |
| Disk Usage     | 30%       | 75%       | 85%       | 100%         |

## Growth Analysis

### Assumptions

- User base grows **20% month-over-month**
- Each user generates **~50 API calls/day**
- Average task creates **3 database records** (task + 2 activity logs)
- Data retention: **1 year**

### 6-Month Projection

| Month | Users | Daily Requests | Storage (PG) | Storage (Mongo) |
| ----- | ----- | -------------- | ------------ | --------------- |
| 1     | 1,000 | 50,000         | 1 GB         | 0.5 GB          |
| 2     | 1,200 | 60,000         | 2 GB         | 1.2 GB          |
| 3     | 1,440 | 72,000         | 3.5 GB       | 2.1 GB          |
| 4     | 1,728 | 86,400         | 5.5 GB       | 3.4 GB          |
| 5     | 2,074 | 103,700        | 8 GB         | 5 GB            |
| 6     | 2,488 | 124,400        | 11 GB        | 7 GB            |

## Scaling Strategy

### Horizontal Scaling (Application)

```
Current:  3 pods  → handles ~300 req/s
Month 3:  5 pods  → handles ~500 req/s
Month 6:  8 pods  → handles ~800 req/s
```

- Managed by HPA with CPU/memory triggers
- Scale-up: +2 pods per 60s, stabilization: 60s
- Scale-down: -1 pod per 120s, stabilization: 300s

### Vertical Scaling (Database)

```
Current:  db-custom-2-4096  (2 vCPU, 4GB)
Month 3:  db-custom-4-8192  (4 vCPU, 8GB)
Month 6:  db-custom-4-16384 (4 vCPU, 16GB) + read replicas
```

### Cache Scaling

```
Current: 1GB Redis Basic
Month 3: 2GB Redis Basic
Month 6: 5GB Redis Standard (with HA)
```

## Cost Estimation (Monthly)

| Resource      | Current  | Month 3  | Month 6  |
| ------------- | -------- | -------- | -------- |
| GKE Nodes     | $140     | $210     | $350     |
| Cloud SQL     | $80      | $160     | $240     |
| Redis         | $35      | $70      | $175     |
| Load Balancer | $20      | $20      | $20      |
| Storage       | $5       | $15      | $30      |
| **Total**     | **$280** | **$475** | **$815** |

## Recommendations

1. **Enable HPA** — Already configured, review thresholds quarterly
2. **Database indexing** — Review slow query logs monthly
3. **Redis eviction** — Configure `maxmemory-policy allkeys-lru`
4. **Read replicas** — Add at Month 4 when read load exceeds 60%
5. **CDN/Caching** — Consider Cloudflare or Cloud CDN at Month 6
6. **Monitoring** — Set up PagerDuty for critical alerts

## Review Schedule

- **Weekly**: Review P95 latency, error rates
- **Monthly**: Review resource utilization, cost analysis
- **Quarterly**: Capacity planning review, growth projection update
