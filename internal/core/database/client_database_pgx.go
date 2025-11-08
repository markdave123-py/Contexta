package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/pgvector/pgvector-go"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/markdave123-py/Contexta/internal/config"
	"github.com/markdave123-py/Contexta/internal/core"
	"github.com/markdave123-py/Contexta/internal/models"
)

type DatabaseClient struct {
	db *sql.DB
}

func NewDatabaseClient(ctx context.Context, cfg *config.Config) (core.DbClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database client configuration is nil")
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is empty")
	}
	if cfg.SslCertPath == "" {
		return nil, fmt.Errorf("SSL_CERT_PATH is empty")
	}
	if _, err := os.Stat(cfg.SslCertPath); err != nil {
		return nil, fmt.Errorf("ssl cert not accessible at %q: %w", cfg.SslCertPath, err)
	}

	// Append SSL params to the provided DATABASE_URL safely.
	u, err := url.Parse(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}
	q := u.Query()
	q.Set("sslmode", "verify-ca")
	q.Set("sslrootcert", cfg.SslCertPath)
	u.RawQuery = q.Encode()
	// dsn := u.String()

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Sensible pool settings for an API service; adjust as needed.
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	// Ensure bootstrap once
	if err := EnsureBootstrapped(ctx, db); err != nil {
		return nil, fmt.Errorf("bootstrap: %w", err)
	}

	return &DatabaseClient{db: db}, nil
}

func (c *DatabaseClient) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Implementing the db interface for user

func (c *DatabaseClient) CreateUser(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("nil user")
	}
	const q = `
		INSERT INTO users (id, first_name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, COALESCE($5, now()), COALESCE($6, now()))
	`
	_, err := c.db.ExecContext(ctx, q,
		user.ID, user.FirstName, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	return err
}

func (c *DatabaseClient) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const q = `
		SELECT id, first_name, email, password_hash, created_at, updated_at
		FROM users WHERE email = $1
	`
	var u models.User
	err := c.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID, &u.FirstName, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Implementing the db interface for Document

func (c *DatabaseClient) CreateDocument(ctx context.Context, doc *models.Document) error {
	if doc == nil {
		return errors.New("nil document")
	}
	const q = `
		INSERT INTO documents
			(id, user_id, file_name, storage_url, source_type, content_type, status, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, COALESCE($8, now()), COALESCE($9, now()))
	`
	_, err := c.db.ExecContext(ctx, q,
		doc.ID, doc.UserID, doc.FileName, doc.StorageURL, doc.SourceType, doc.ContentType, doc.Status, doc.CreatedAt, doc.UpdatedAt)
	return err
}

func (c *DatabaseClient) GetDocumentByID(ctx context.Context, id string) (*models.Document, error) {
	const q = `
		SELECT id, user_id, file_name, storage_url, source_type, content_type, status, created_at, updated_at
		FROM documents
		WHERE id = $1
	`
	var d models.Document
	err := c.db.QueryRowContext(ctx, q, id).Scan(
		&d.ID, &d.UserID, &d.FileName, &d.StorageURL, &d.SourceType, &d.ContentType, &d.Status, &d.CreatedAt, &d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (c *DatabaseClient) ListDocumentsByUser(ctx context.Context, userID string) ([]models.Document, error) {
	const q = `
		SELECT id, user_id, file_name, storage_url, source_type, status, created_at, updated_at
		FROM documents
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := c.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Document
	for rows.Next() {
		var d models.Document
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.FileName, &d.StorageURL, &d.SourceType, &d.Status, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (c *DatabaseClient) UpdateDocumentStatus(ctx context.Context, id string, status string) error {
	const q = `
		UPDATE documents
		SET status = $2, updated_at = now()
		WHERE id = $1
	`
	res, err := c.db.ExecContext(ctx, q, id, status)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("document not found: %s", id)
	}
	return nil
}

// // Implementing the db interface for Document Chunks

// InsertDocumentChunks inserts chunks in a single transaction.
func (c *DatabaseClient) InsertDocumentChunks(ctx context.Context, chunks []models.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	const q = `
		INSERT INTO document_chunks
			(id, document_id, position, text, embedding, token_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, now()))
	`
	stmt, err := tx.PrepareContext(ctx, q)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	for i := range chunks {
		ch := &chunks[i]
		vec := pgvector.NewVector(ch.Embedding)

		// Embedding []float32 maps to pgvector via pgx stdlib; ensure your pgx/stdlib is imported.
		if _, err := stmt.ExecContext(ctx,
			ch.ID, ch.DocumentID, ch.Position, ch.Text, vec, ch.TokenCount, ch.CreatedAt,
		); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (c *DatabaseClient) GetChunksByDocument(ctx context.Context, documentID string) ([]models.DocumentChunk, error) {
	const q = `
		SELECT id, document_id, position, text, embedding, token_count, created_at
		FROM document_chunks
		WHERE document_id = $1
		ORDER BY position ASC
	`
	rows, err := c.db.QueryContext(ctx, q, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.DocumentChunk
	for rows.Next() {
		var ch models.DocumentChunk
		if err := rows.Scan(
			&ch.ID, &ch.DocumentID, &ch.Position, &ch.Text, &ch.Embedding, &ch.TokenCount, &ch.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

// SearchDocumentChunks finds top-k similar chunks within a document for a query embedding.
func (c *DatabaseClient) SearchDocumentChunks(ctx context.Context, docID string, queryVec []float32, limit int) ([]models.DocumentChunk, error) {
    const q = `
        SELECT id, document_id, position, text, embedding, token_count
        FROM document_chunks
        WHERE document_id = $1
        ORDER BY embedding <-> $2
        LIMIT $3
    `
	vec := pgvector.NewVector(queryVec)
    rows, err := c.db.QueryContext(ctx, q, docID, vec, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []models.DocumentChunk
    for rows.Next() {
        var (
            ch  models.DocumentChunk
            emb pgvector.Vector
        )
        if err := rows.Scan(&ch.ID, &ch.DocumentID, &ch.Position, &ch.Text, &emb, &ch.TokenCount); err != nil {
            return nil, err
        }
		ch.Embedding = emb.Slice()
        out = append(out, ch)
    }
    return out, nil
}

