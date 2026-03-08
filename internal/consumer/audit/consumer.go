package audit

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/matheusm25/audit-ingestion-service/internal/model"
	"github.com/matheusm25/audit-ingestion-service/internal/platform/rabbitmq"
	"github.com/matheusm25/audit-ingestion-service/internal/repository"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Listener struct {
	rabbitmqConnection rabbitmq.RabbitMQConnection
	ctx                context.Context
	consumerConfig     ConsumerConfig
	auditRepository    *repository.AuditRepository
	batch              []model.AuditEvent
	batchMutex         sync.Mutex
	lastFlushTime      time.Time
}

type ConsumerConfig struct {
	QueueName          string
	BatchIngestionSize int
	BatchFlushInterval int
	PrefetchCount      int
}

func NewListener(
	rabbitmqConnection rabbitmq.RabbitMQConnection,
	ctx context.Context,
	consumerConfig ConsumerConfig,
	auditRepository *repository.AuditRepository,
) *Listener {
	return &Listener{
		rabbitmqConnection: rabbitmqConnection,
		ctx:                ctx,
		consumerConfig:     consumerConfig,
		auditRepository:    auditRepository,
		batch:              make([]model.AuditEvent, 0, consumerConfig.BatchIngestionSize),
		lastFlushTime:      time.Now(),
	}
}

func (l *Listener) ListenForMessages() {
	messages, err := l.rabbitmqConnection.Subscribe(l.consumerConfig.QueueName, l.consumerConfig.PrefetchCount, true)
	if err != nil {
		log.Fatal("Failed to subscribe to RabbitMQ queue:", err)
	}

	flushTicker := time.NewTicker(time.Duration(l.consumerConfig.BatchFlushInterval) * time.Second)
	defer flushTicker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			log.Println("shutting down consumer")
			l.flushBatch()
			return

		case <-flushTicker.C:
			l.flushBatch()

		case msg, ok := <-messages:
			if !ok {
				log.Println("message channel closed")
				l.flushBatch()
				return
			}

			l.processMessage(msg)
		}
	}
}

func (l *Listener) processMessage(msg amqp.Delivery) {
	parsedMessage := model.AuditEvent{}

	if err := json.Unmarshal(msg.Body, &parsedMessage); err != nil {
		log.Printf("Failed to unmarshal message: %v, error: %v", string(msg.Body), err)
		msg.Nack(false, false)
		return
	}

	if err := parsedMessage.Validate(); err != nil {
		log.Printf("Invalid message: %v, error: %v", parsedMessage, err)
		msg.Nack(false, false)
		return
	}

	l.batchMutex.Lock()
	l.batch = append(l.batch, parsedMessage)
	batchSize := len(l.batch)
	l.batchMutex.Unlock()

	msg.Ack(false)

	if batchSize >= l.consumerConfig.BatchIngestionSize {
		l.flushBatch()
	}
}

func (l *Listener) flushBatch() {
	l.batchMutex.Lock()
	defer l.batchMutex.Unlock()

	if len(l.batch) == 0 {
		return
	}

	log.Printf("Flushing batch with %d events", len(l.batch))

	if err := l.auditRepository.InsertBatch(l.ctx, l.batch); err != nil {
		log.Printf("Failed to insert batch: %v", err)
		return
	}

	log.Printf("Successfully inserted %d events", len(l.batch))

	l.batch = make([]model.AuditEvent, 0, l.consumerConfig.BatchIngestionSize)
	l.lastFlushTime = time.Now()
}
