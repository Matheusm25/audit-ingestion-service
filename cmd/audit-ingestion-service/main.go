package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/matheusm25/audit-ingestion-service/internal/config"
	"github.com/matheusm25/audit-ingestion-service/internal/consumer/audit"
	"github.com/matheusm25/audit-ingestion-service/internal/http/server"
	"github.com/matheusm25/audit-ingestion-service/internal/platform/clickhouse"
	"github.com/matheusm25/audit-ingestion-service/internal/platform/rabbitmq"
	"github.com/matheusm25/audit-ingestion-service/internal/repository"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	} else {
		fmt.Println("Environment variables loaded!")
	}

	rabbitmqConnection, err := rabbitmq.NewConnection(rabbitmq.RabbitMQConfig{
		ConnectionUrl: cfg.RabbitMQ.ConnectionUrl,
	})
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	} else {
		fmt.Println("RabbitMQ connected!")
		defer rabbitmqConnection.Close()
	}

	clickhouseConn, err := clickhouse.NewConnection(clickhouse.ClickHouseConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		Username: cfg.Database.Username,
		Password: cfg.Database.Password,
		Database: cfg.Database.Database,
	})
	if err != nil {
		log.Fatal("Failed to connect to ClickHouse:", err)
	} else {
		fmt.Println("Clickhouse connected!")
		defer clickhouseConn.Close()
	}

	auditRepo := repository.NewAuditRepository(clickhouseConn)

	httpServer := server.NewServer(cfg.App.HTTPPort)
	go func() {
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start HTTP server:", err)
		}
	}()

	listener := audit.NewListener(
		rabbitmqConnection,
		ctx,
		audit.ConsumerConfig{
			QueueName:          cfg.RabbitMQ.IngestionQueue,
			BatchIngestionSize: cfg.App.BatchIngestionSize,
			BatchFlushInterval: cfg.App.BatchFlushInterval,
			PrefetchCount:      cfg.App.BatchIngestionSize,
		},
		auditRepo,
	)

	log.Println("Starting audit ingestion service...")
	listener.ListenForMessages()

	if err := httpServer.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}
