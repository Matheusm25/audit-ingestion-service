package repository

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/matheusm25/audit-ingestion-service/internal/model"
)

type AuditRepository struct {
	conn clickhouse.Conn
}

func NewAuditRepository(conn clickhouse.Conn) *AuditRepository {
	return &AuditRepository{
		conn: conn,
	}
}

func (r *AuditRepository) InsertBatch(ctx context.Context, events []model.AuditEvent) error {
	batch, err := r.conn.PrepareBatch(ctx, `
		INSERT INTO audit_log
		(timestamp, entity, entity_id, action, user_id, payload) VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}

	for _, event := range events {
		err := batch.Append(
			event.Timestamp,
			event.Entity,
			event.EntityID,
			event.Action,
			event.UserID,
			string(event.Payload),
		)

		if err != nil {
			return err
		}
	}

	return batch.Send()
}
