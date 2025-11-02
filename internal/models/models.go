package models

import (
	"time"
)

// User represents an authenticated user of the system.
type User struct {
	ID           string    `db:"id" json:"id"`
	FirstName    string    `db:"first_name" json:"first_name"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// Document represents a user-uploaded or crawled document.
type Document struct {
	ID          string    `db:"id" json:"id"`
	UserID      string    `db:"user_id" json:"user_id"`
	FileName    string    `db:"file_name" json:"file_name"`
	StorageURL  string    `db:"storage_url" json:"storage_url"` // S3 URL or original link
	SourceType  string    `db:"source_type" json:"source_type"` // "upload" or "url"
	Status      string    `db:"status" json:"status"`           // uploaded | processing | ready | failed
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// DocumentChunk represents one text chunk from a document.
type DocumentChunk struct {
	ID          string    `db:"id" json:"id"`
	DocumentID  string    `db:"document_id" json:"document_id"`
	Text        string    `db:"text" json:"text"`
	Embedding   []float32 `db:"embedding" json:"embedding"` // pgvector column
	Position    int       `db:"position" json:"position"`
	TokenCount  int       `db:"token_count" json:"token_count"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// ChatSession represents one conversation session for a document.
type ChatSession struct {
	ID          string    `db:"id" json:"id"`
	UserID      string    `db:"user_id" json:"user_id"`
	DocumentID  string    `db:"document_id" json:"document_id"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// ChatMessage represents an individual chat message (user or assistant).
type ChatMessage struct {
	ID         string    `db:"id" json:"id"`
	SessionID  string    `db:"session_id" json:"session_id"`
	Role       string    `db:"role" json:"role"`       // "user" or "assistant"
	Content    string    `db:"content" json:"content"` // message text
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
