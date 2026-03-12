package repository

import (
	"context"
	"strings"

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

type AuditFilter struct {
	UserID   string
	Entity   string
	EntityID string
	Action   string
	OrderBy  string
	Page     int
	PerPage  int
}

func (r *AuditRepository) ListAudits(ctx context.Context, filter AuditFilter) ([]model.AuditEvent, error) {
	query := "SELECT timestamp, entity, entity_id, action, user_id, payload FROM audit_log"

	var conditions []string
	var args []interface{}

	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}

	if filter.Entity != "" {
		conditions = append(conditions, "entity = ?")
		args = append(args, filter.Entity)
	}

	if filter.EntityID != "" {
		conditions = append(conditions, "entity_id = ?")
		args = append(args, filter.EntityID)
	}

	if filter.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, filter.Action)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if filter.OrderBy == "asc" {
		query += " ORDER BY timestamp ASC"
	} else {
		query += " ORDER BY timestamp DESC"
	}

	perPage := filter.PerPage
	if perPage <= 0 || perPage > 1000 {
		perPage = 100
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}

	limit := perPage
	offset := (page - 1) * perPage

	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.AuditEvent
	for rows.Next() {
		var event model.AuditEvent
		var payloadStr string

		err := rows.Scan(
			&event.Timestamp,
			&event.Entity,
			&event.EntityID,
			&event.Action,
			&event.UserID,
			&payloadStr,
		)
		if err != nil {
			return nil, err
		}

		event.Payload = []byte(payloadStr)
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *AuditRepository) CountAudits(ctx context.Context, filter AuditFilter) (uint64, error) {
	query := "SELECT COUNT(*) FROM audit_log"

	var conditions []string
	var args []interface{}

	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}

	if filter.Entity != "" {
		conditions = append(conditions, "entity = ?")
		args = append(args, filter.Entity)
	}

	if filter.EntityID != "" {
		conditions = append(conditions, "entity_id = ?")
		args = append(args, filter.EntityID)
	}

	if filter.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, filter.Action)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count uint64
	err := r.conn.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
