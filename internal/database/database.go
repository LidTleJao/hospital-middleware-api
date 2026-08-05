// Package database opens and verifies the connection pool the repositories share.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	// Registers the "pgx" driver with database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open connects to Postgres and verifies the connection before returning it,
// so a bad DSN or an unreachable server fails at boot rather than on the
// first request.
func Open(ctx context.Context, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// A middleware in front of one HIS does not need a large pool; these
	// bounds keep idle connections from being reaped by Postgres and cap
	// what a burst of requests can hold open.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
