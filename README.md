# Demo Services

Three Go services built from one image for the observable Task Manager demo:

- `server`: task API and checkout entry point
- `inventory`: checkout orchestration, PostgreSQL persistence, and Redis state
- `pricing`: quote service with an external market-data dependency

## Features

- **PostgreSQL** for persistent task storage with auto-migration
- **Redis** for read-through caching (optional, degrades gracefully)
- **Kafka** for async audit event streaming (optional, degrades gracefully)
- **Cross-service checkout** from the API through inventory and pricing
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
| POST | `/api/checkout` | Run the checkout path through inventory and pricing |

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `REDIS_URL` | No | Redis connection string |
| `KAFKA_BROKERS` | No | Kafka broker addresses |
| `KAFKA_TOPIC` | No | Kafka topic (default: `task-events`) |
| `PORT` | No | Server port (default: `8080`) |
| `INVENTORY_URL` | No | Inventory service URL; enables checkout |
| `PRICING_URL` | No | Pricing service URL for the inventory service |
| `MARKET_DATA_URL` | No | External quote provider used by pricing |

The inventory and pricing binaries use the same image:

```yaml
command: ["/inventory"]
```

```yaml
command: ["/pricing"]
```

<!-- ci canary: verify reusable workflow after action major bumps 2026-11-08 -->
