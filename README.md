# Demo API

Production-ready Go REST API for the Task Manager demo.

## Features

- **PostgreSQL** for persistent task storage with auto-migration
- **Redis** for read-through caching (optional, degrades gracefully)
- **Kafka** for async audit event streaming (optional, degrades gracefully)
- **Structured JSON logging** via `slog`
- **Graceful shutdown** with signal handling
- **Distroless container** for minimal attack surface
- **Health/readiness probes** for Kubernetes

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Liveness probe |
| GET | `/readyz` | Readiness probe (checks DB) |
| GET | `/api/tasks` | List all tasks |
| POST | `/api/tasks` | Create a task |
| GET | `/api/tasks/:id` | Get a task (cached) |
| PUT | `/api/tasks/:id` | Update a task |
| DELETE | `/api/tasks/:id` | Delete a task |

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `REDIS_URL` | No | Redis connection string |
| `KAFKA_BROKERS` | No | Kafka broker addresses |
| `KAFKA_TOPIC` | No | Kafka topic (default: `task-events`) |
| `PORT` | No | Server port (default: `8080`) |
