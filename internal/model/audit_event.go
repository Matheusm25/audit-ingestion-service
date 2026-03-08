package model

import (
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
)

type AuditEvent struct {
	Timestamp time.Time       `json:"timestamp" validate:"required"`
	Entity    string          `json:"entity" validate:"required"`
	EntityID  string          `json:"entityId" validate:"required"`
	Action    string          `json:"action" validate:"required"`
	UserID    string          `json:"userId" validate:"required"`
	Payload   json.RawMessage `json:"payload" validate:"required"`
}

var validate = validator.New()

func (e *AuditEvent) Validate() error {
	return validate.Struct(e)
}
