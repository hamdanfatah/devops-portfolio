# System Architecture

## High-Level Architecture

```mermaid
flowchart TB
    subgraph Client
        USER[User / API Client]
    end

    subgraph "Edge Layer"
        LB[Cloud Load Balancer]
        ING[Nginx Ingress Controller]
    end

    subgraph "Application Layer"
        API1[API Pod 1]
        API2[API Pod 2]
        API3[API Pod 3]
    end

    subgraph "Data Layer"
        PG[(PostgreSQL\nPrimary Store)]
        MDB[(MongoDB\nActivity Logs)]
        RD[(Redis\nCache Layer)]
    end

    subgraph "Monitoring"
        PROM[Prometheus]
        GRAF[Grafana]
        ALERT[Alertmanager]
    end

    USER --> LB --> ING
    ING --> API1 & API2 & API3
    API1 & API2 & API3 --> RD
    API1 & API2 & API3 --> PG
    API1 & API2 & API3 --> MDB
    API1 & API2 & API3 -.->|metrics| PROM
    PROM --> GRAF
    PROM --> ALERT
```

## Request Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant API as Go API
    participant R as Redis
    participant PG as PostgreSQL
    participant M as MongoDB

    C->>API: GET /api/tasks/:id
    API->>R: Check cache
    alt Cache Hit
        R-->>API: Return cached task
    else Cache Miss
        API->>PG: Query task
        PG-->>API: Return task
        API->>R: Cache task (TTL: 5m)
    end
    API-->>C: JSON response

    C->>API: POST /api/tasks
    API->>PG: Insert task
    PG-->>API: Return created task
    API->>R: Cache new task
    API-->>M: Log activity (async)
    API-->>C: 201 Created
```

## Deployment Architecture

```mermaid
graph TB
    subgraph "GCP - asia-southeast2"
        subgraph "VPC Network (10.0.0.0/20)"
            subgraph "GKE Cluster"
                NP[Node Pool\n2-10 nodes\ne2-standard-2]
                subgraph "task-manager namespace"
                    D[Deployment\n3 replicas]
                    HPA[HPA\nmin:2 max:20]
                    SVC[ClusterIP Service]
                    ING2[Ingress + TLS]
                end
            end
            SQL[Cloud SQL PostgreSQL\nHA - Regional]
            MEM[Memorystore Redis\n1GB]
        end
        AR[Artifact Registry]
        GCS[Cloud Storage\nBackups]
    end

    subgraph "CI/CD"
        GHA[GitHub Actions]
    end

    GHA -->|push image| AR
    AR -->|pull image| D
    HPA -->|scale| D
    D --> SQL
    D --> MEM
```

## Design Decisions

### Why Golang?

- Compiled binary → minimal Docker image (<30MB)
- Built-in concurrency with goroutines
- Ecosystem: Docker, Kubernetes, Terraform all written in Go
- Strong typing reduces production bugs

### Why Multi-Database?

- **PostgreSQL**: ACID for critical task data, strong query capabilities
- **MongoDB**: Flexible schema for audit/activity logs, easy aggregation
- **Redis**: Sub-millisecond caching, reduces DB load by ~70%

### Why Cache-Aside Pattern?

- Application controls cache lifecycle
- Database remains source of truth
- Graceful degradation if Redis is down (fallback to DB)

### High Availability Strategies

1. **HPA**: Auto-scale pods based on CPU/memory metrics
2. **Topology Spread**: Distribute pods across nodes
3. **Rolling Updates**: Zero-downtime deployments
4. **Health Probes**: Startup, liveness, readiness checks
5. **Cloud SQL HA**: Regional instance with automatic failover
