package clickhouse

import (
	"context"
	"strconv"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

func NewConnection(cfg ClickHouseConfig) (clickhouse.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{
			cfg.Host + ":" + strconv.Itoa(cfg.Port),
		},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		DialTimeout:     time.Second * 5,
	})

	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	return conn, nil
}
