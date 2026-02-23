# Task Manager API

A production-ready REST API built with **Go** (Gin framework) that demonstrates multi-database integration with **PostgreSQL**, **MongoDB**, and **Redis**.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Gin Router                        │
│              (middleware stack)                       │
├─────────────────────────────────────────────────────┤
│                  Handler Layer                       │
│           (request/response handling)                │
├─────────────────────────────────────────────────────┤
│                  Service Layer                       │
│          (business logic, cache-aside)               │
├──────────┬──────────────┬───────────────────────────┤
│ PostgreSQL│   MongoDB    │         Redis              │
│ (Tasks)   │ (Activity    │   (Cache Layer)            │
│           │   Logs)      │                            │
└──────────┴──────────────┴───────────────────────────┘
```

## Quick Start

```bash
# Start all services
docker compose up -d

# API is available at http://localhost:8080
# Health check: http://localhost:8080/health
```

## API Endpoints

| Method | Endpoint                    | Description         |
| ------ | --------------------------- | ------------------- |
| GET    | `/health`                   | Health check        |
| POST   | `/api/tasks`                | Create a task       |
| GET    | `/api/tasks`                | List tasks          |
| GET    | `/api/tasks/:id`            | Get task by ID      |
| PUT    | `/api/tasks/:id`            | Update a task       |
| DELETE | `/api/tasks/:id`            | Delete a task       |
| GET    | `/api/tasks/:id/activities` | Get task activities |

## Example Requests

```bash
# Create a task
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Deploy to production", "description": "Deploy v1.0", "priority": "high"}'

# List tasks
curl http://localhost:8080/api/tasks?page=1&per_page=10&status=pending

# Update a task
curl -X PUT http://localhost:8080/api/tasks/<id> \
  -H "Content-Type: application/json" \
  -d '{"status": "completed"}'
```

## Tech Stack

- **Go 1.22** with Gin framework
- **PostgreSQL 16** — Primary data store
- **MongoDB 7.0** — Activity logging
- **Redis 7** — Caching layer (cache-aside pattern)
- **Docker** — Multi-stage build, non-root user
- **Zap** — Structured logging
