package db

import (
	"context"

	"github.com/markdave123-py/Contexta/internal/models"
)

// DbClient defines all persistence operations your services will need.
// It abstracts Postgres/pgvector so higher layers never depend on a specific DB.
type DbClient interface {
	CreateUser(ctx context.Context, user *models.User) (err error)
	GetUserByEmail(ctx context.Context, email string) (user *models.User, err error)

	CreateDocument(ctx context.Context, doc *models.Document) error
	GetDocumentByID(ctx context.Context, id string) (*models.Document, error)
	ListDocumentsByUser(ctx context.Context, userID string) ([]models.Document, error)
	UpdateDocumentStatus(ctx context.Context, id string, status string) error

	InsertDocumentChunks(ctx context.Context, chunks []models.DocumentChunk) error
	GetChunksByDocument(ctx context.Context, documentID string) ([]models.DocumentChunk, error)

	SearchDocumentChunks(ctx context.Context, docID string, queryVec []float32, limit int) ([]models.DocumentChunk, error)

	Close() error

	// CreateChatSession(ctx context.Context, session *models.ChatSession) error
	// AddChatMessage(ctx context.Context, message *models.ChatMessage) error
	// GetMessagesBySession(ctx context.Context, sessionID string) ([]models.ChatMessage, error)
}
