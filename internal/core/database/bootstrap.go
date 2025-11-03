// internal/infra/db/bootstrap.go
package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"
)

//go:embed scripts/initdb.sql

var bootstrapFS embed.FS

func EnsureBootstrapped(ctx context.Context, db *sql.DB) error {

	ctxBoot, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	var exists bool
	err := db.QueryRowContext(ctxBoot, `
		SELECT EXISTS (
		  SELECT 1 FROM information_schema.tables 
		  WHERE table_name = 'contexta_meta'
		)`).
		Scan(&exists)
	if err != nil {
		return fmt.Errorf("meta table check failed: %w", err)
	}

	if exists{
		println("meta data already exists!")
		// return nil
	}

	// 2) If table missing OR version row missing, run bootstrap.sql
	if !exists {
		return runBootstrap(ctxBoot, db)
	}

	var hasVersion bool
	if err := db.QueryRowContext(ctxBoot, `SELECT EXISTS (SELECT 1 FROM contexta_meta WHERE version = 1)`).Scan(&hasVersion); err != nil {
		return fmt.Errorf("meta version check failed: %w", err)
	}
	if !hasVersion {
		return runBootstrap(ctxBoot, db)
	}

	return nil
}

func runBootstrap(ctx context.Context, db *sql.DB) error {
	sqlBytes, err := bootstrapFS.ReadFile("scripts/initdb.sql")
	if err != nil {
		return fmt.Errorf("read initdb.sql: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("exec bootstrap: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit bootstrap: %w", err)
	}
	return nil
}
