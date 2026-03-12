# 🔍 Audit Ingestion Service

A high-performance Go service that consumes audit events from RabbitMQ and efficiently stores them in ClickHouse using batch processing. Built for scalability and reliability with Dead Letter Queue (DLQ) support for fault tolerance.

## ✨ Features

- **Real-time Event Ingestion**: Listens to RabbitMQ queues for audit events
- **Batch Processing**: Optimized batch insertion to ClickHouse for improved performance
- **Dead Letter Queue (DLQ)**: Automatic handling of failed messages with DLQ implementation
- **Configurable Batching**: Flexible batch size and flush interval configuration
- **Graceful Shutdown**: Ensures all pending batches are processed before shutdown
- **Message Validation**: Validates incoming audit events before processing
- **Environment-based Configuration**: Easy configuration through environment variables

## 🏗️ Architecture

```
┌─────────────┐         ┌──────────────────────┐         ┌─────────────┐
│  RabbitMQ   │────────▶│  Audit Ingestion     │────────▶│ ClickHouse  │
│   Queue     │         │     Service          │         │  Database   │
└─────────────┘         └──────────────────────┘         └─────────────┘
                                │
                                ▼
                        ┌──────────────────┐
                        │  Dead Letter     │
                        │     Queue        │
                        └──────────────────┘
```

The service:
1. Subscribes to a RabbitMQ queue
2. Validates and accumulates messages in memory batches
3. Flushes batches to ClickHouse based on:
   - Batch size threshold (`BATCH_INGESTION_SIZE`)
   - Time interval (`BATCH_FLUSH_INTERVAL_IN_SECONDS`)
4. Routes failed messages to a Dead Letter Queue for later inspection

## 📋 Prerequisites

### Option 1: Running Locally
- **Go**: 1.25.7 or higher
- **RabbitMQ**: Running instance
- **ClickHouse**: Running instance with appropriate database and table

### Option 2: Running with Docker
- **Docker**: 20.10 or higher
- **RabbitMQ**: Running instance (local or remote)
- **ClickHouse**: Running instance (local or remote)

### ClickHouse Table Schema

Create the following table in your ClickHouse database:

```sql
CREATE TABLE audit_log
(
    timestamp DateTime,
    user_id String,
    action LowCardinality(String),
    resource LowCardinality(String),
    ip String,
    status_code UInt16,
    metadata String
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (user_id, timestamp)
TTL timestamp + INTERVAL 12 MONTH
SETTINGS index_granularity = 8192;
```

## 🚀 Getting Started

### Installation

```bash
# Clone the repository
git clone https://github.com/matheusm25/audit-ingestion-service.git
cd audit-ingestion-service

# Install dependencies
go mod download

# Build the application
go build -o audit-ingestion-service ./cmd/audit-ingestion-service
```

### Configuration

Create a `.env` file in the root directory (or set environment variables):

```bash
cp .env.example .env
```

### Running the Service

#### Option 1: Run Locally

```bash
# Using the built binary
./audit-ingestion-service

# Or run directly with Go
go run ./cmd/audit-ingestion-service/main.go
```

#### Option 2: Run with Docker

```bash
# Build the Docker image
docker build -t audit-ingestion-service .

# Run the container (connecting to host services)
docker run --rm \
  -e RABBITMQ_CONNECTION_URL=amqp://localhost:5672/ \
  -e RABBITMQ_INGESTION_QUEUE_NAME=audit-ingestion \
  -e CLICKHOUSE_HOST=localhost \
  -e CLICKHOUSE_PORT=9000 \
  -e CLICKHOUSE_DATABASE=default \
  -e CLICKHOUSE_USERNAME=default \
  -e CLICKHOUSE_PASSWORD= \
  -e BATCH_INGESTION_SIZE=100 \
  -e BATCH_FLUSH_INTERVAL_IN_SECONDS=30 \
  --network host \
  audit-ingestion-service
```

**Note**: The `--network host` flag allows the container to connect to services running on your host machine.

## ⚙️ Environment Variables

| Variable | Description | Default Value |
|----------|-------------|---------------|
| `RABBITMQ_CONNECTION_URL` | RabbitMQ connection URL (AMQP format) | `amqp://localhost` |
| `RABBITMQ_INGESTION_QUEUE_NAME` | Name of the queue to consume audit events from | `audit-ingestion` |
| `CLICKHOUSE_HOST` | ClickHouse server hostname | `localhost` |
| `CLICKHOUSE_PORT` | ClickHouse server port | `9000` |
| `CLICKHOUSE_DATABASE` | ClickHouse database name | `audit-ingestion-service` |
| `CLICKHOUSE_USERNAME` | ClickHouse username | `default` |
| `CLICKHOUSE_PASSWORD` | ClickHouse password | `""` (empty) |
| `BATCH_INGESTION_SIZE` | Number of events to accumulate before flushing to ClickHouse | `100` |
| `BATCH_FLUSH_INTERVAL_IN_SECONDS` | Maximum time (in seconds) to wait before flushing a batch | `30` |
| `HTTP_PORT` | Port for the HTTP server (health check endpoint) | `8080` |

### Configuration Tips

- **Batch Size**: Increase `BATCH_INGESTION_SIZE` for higher throughput with more memory usage
- **Flush Interval**: Lower `BATCH_FLUSH_INTERVAL_IN_SECONDS` for more real-time processing
- **Prefetch Count**: The service sets RabbitMQ prefetch count equal to `BATCH_INGESTION_SIZE` for optimal performance

## 📊 Audit Event Format

Messages published to the RabbitMQ queue must follow this JSON structure:

```json
{
  "timestamp": "2026-03-08T10:30:00Z",
  "entity": "user",
  "entityId": "user-123",
  "action": "created",
  "userId": "admin-456",
  "payload": {
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

### Field Descriptions

- **timestamp** *(required)*: ISO 8601 timestamp of when the event occurred
- **entity** *(required)*: The type of entity affected (e.g., user, order, product)
- **entityId** *(required)*: Unique identifier of the affected entity
- **action** *(required)*: The action performed (e.g., created, updated, deleted)
- **userId** *(required)*: Identifier of the user who performed the action
- **payload** *(required)*: JSON object containing additional event data

## 🛡️ Dead Letter Queue (DLQ)

The service implements a Dead Letter Queue pattern for handling message processing failures:

- **DLQ Exchange**: `{queue-name}_dlx`
- **DLQ Queue**: `{queue-name}_dlq`

Messages are routed to the DLQ when:
- JSON deserialization fails
- Message validation fails

Failed messages in the DLQ can be inspected, corrected, and re-queued manually.

## 🐳 Docker Support

The service can be run in a Docker container. The Dockerfile uses a multi-stage build to create a minimal Alpine-based image.

### Build

```bash
docker build -t audit-ingestion-service .
```

### Run

```bash
# Basic run with environment variables
docker run --rm \
  -e RABBITMQ_CONNECTION_URL=amqp://your-rabbitmq:5672/ \
  -e RABBITMQ_INGESTION_QUEUE_NAME=audit-ingestion \
  -e CLICKHOUSE_HOST=your-clickhouse \
  -e CLICKHOUSE_PORT=9000 \
  -e CLICKHOUSE_DATABASE=default \
  -e CLICKHOUSE_USERNAME=default \
  -e CLICKHOUSE_PASSWORD=yourpassword \
  audit-ingestion-service

# Or use --network host to connect to local services
docker run --rm --network host \
  -e RABBITMQ_CONNECTION_URL=amqp://localhost:5672/ \
  -e CLICKHOUSE_HOST=localhost \
  audit-ingestion-service
```

### Environment File

Alternatively, use an environment file:

```bash
# Create .env.docker with your configuration
docker run --rm --env-file .env.docker audit-ingestion-service
```

## 🏥 Health Check Endpoints

The service exposes HTTP health check endpoints for monitoring:

### Basic Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy"
}
```

### Services Health Check

Checks connectivity with RabbitMQ and ClickHouse:

```bash
curl http://localhost:8080/health/services
```

Response when all services are healthy (HTTP 200):
```json
{
  "rabbitMQStatus": "healthy",
  "clickHouseStatus": "healthy"
}
```

Response when any service is unhealthy (HTTP 503):
```json
{
  "rabbitMQStatus": "unhealthy",
  "clickHouseStatus": "healthy"
}
```

## 📊 Query Audit Data

The service provides an HTTP endpoint to query and filter audit data from ClickHouse.

### List Audits

```bash
GET /audits
```

#### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `userId` | string | Filter by user ID |
| `entity` | string | Filter by entity type (e.g., user, order) |
| `entityId` | string | Filter by entity ID |
| `action` | string | Filter by action (e.g., created, updated, deleted) |
| `orderBy` | string | Order by timestamp: `asc` or `desc` (default: `desc`) |
| `page` | integer | Page number (default: `1`) |
| `perPage` | integer | Items per page (default: `100`, max: `1000`) |

#### Examples

Get all audits (first page, 100 items):
```bash
curl "http://localhost:8080/audits"
```

Filter by user ID:
```bash
curl "http://localhost:8080/audits?userId=user-123"
```

Filter by entity and action:
```bash
curl "http://localhost:8080/audits?entity=order&action=created"
```

Pagination with custom page size:
```bash
curl "http://localhost:8080/audits?page=2&perPage=50"
```

Combined filters with sorting:
```bash
curl "http://localhost:8080/audits?userId=admin-456&entity=user&orderBy=asc&page=1&perPage=20"
```

#### Response

```json
{
  "audits": [
    {
      "timestamp": "2026-03-08T10:30:00Z",
      "entity": "user",
      "entityId": "user-123",
      "action": "created",
      "userId": "admin-456",
      "payload": {
        "email": "user@example.com",
        "name": "John Doe"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "perPage": 100,
    "totalPages": 5,
    "totalCount": 450,
    "count": 100
  }
}
```

## �️ Roadmap

The following features and improvements are planned for future releases:

- [ ] **Database Migrations**: Create migration scripts to automatically initialize the ClickHouse audit table schema
- [ ] **Docker Hub Publishing**: Publish official Docker images to Docker Hub for easy distribution
## �📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the issues page.
