package health

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/matheusm25/audit-ingestion-service/internal/platform/rabbitmq"
)

const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
)

type Checker struct {
	rabbitmqConn   *rabbitmq.RabbitMQConnection
	clickhouseConn clickhouse.Conn
}

func NewChecker(rabbitmqConn *rabbitmq.RabbitMQConnection, clickhouseConn clickhouse.Conn) *Checker {
	return &Checker{
		rabbitmqConn:   rabbitmqConn,
		clickhouseConn: clickhouseConn,
	}
}

func (c *Checker) CheckRabbitMQ() string {
	if c.rabbitmqConn == nil || c.rabbitmqConn.Conn == nil {
		return StatusUnhealthy
	}

	if c.rabbitmqConn.Conn.IsClosed() {
		return StatusUnhealthy
	}

	return StatusHealthy
}

func (c *Checker) CheckClickHouse() string {
	if c.clickhouseConn == nil {
		return StatusUnhealthy
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := c.clickhouseConn.Ping(ctx); err != nil {
		return StatusUnhealthy
	}

	return StatusHealthy
}
