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

- **Go**: 1.25.7 or higher
- **RabbitMQ**: Running instance
- **ClickHouse**: Running instance with appropriate database and table

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

```bash
# Using the built binary
./audit-ingestion-service

# Or run directly with Go
go run ./cmd/audit-ingestion-service/main.go
```

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

> **Note**: Dockerfile is currently under development and will be added soon. This will enable easy containerized deployment of the service.

## �️ Roadmap

The following features and improvements are planned for future releases:

- [ ] **Database Migrations**: Create migration scripts to automatically initialize the ClickHouse audit table schema
- [ ] **Docker Support**: Develop Dockerfile for containerized deployment
- [ ] **Docker Hub Publishing**: Publish official Docker images to Docker Hub for easy distribution

## �📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the issues page.
