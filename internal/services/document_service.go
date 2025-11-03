package services

import (
	"context"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/markdave123-py/Contexta/internal/core"
	"github.com/markdave123-py/Contexta/internal/models"
)

type DocumentService struct {
	db      core.DbClient
	storage core.ObjectClient
	bucket  string
}

func NewDocumentService(db core.DbClient, storage core.ObjectClient, bucket string) *DocumentService {
	return &DocumentService{db: db, storage: storage, bucket: bucket}
}

func (s *DocumentService) UploadAndCreate(ctx context.Context, userID, filename, contentType string, data []byte, sourceType string) (*models.Document, error) {
	docID := uuid.NewString()
	key := s.objectKey(userID, docID, filename)

	url, err := s.storage.UploadFile(ctx, s.bucket, key, data, contentType)
	if err != nil {
		return nil, err
	}

	doc := &models.Document{
		ID:         docID,
		UserID:     userID,
		FileName:   filename,
		StorageURL: url,
		SourceType: sourceType, // "upload" or "url"
		Status:     "uploaded",
	}
	if err := s.db.CreateDocument(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *DocumentService) Get(ctx context.Context, id string) (*models.Document, error) {
	return s.db.GetDocumentByID(ctx, id)
}

func (s *DocumentService) ListByUser(ctx context.Context, userID string) ([]models.Document, error) {
	return s.db.ListDocumentsByUser(ctx, userID)
}

func (s *DocumentService) SetStatus(ctx context.Context, docID string, status string) error {
	return s.db.UpdateDocumentStatus(ctx, docID, status)
}

func (s *DocumentService) InsertChunks(ctx context.Context, chunks []models.DocumentChunk) error {
	return s.db.InsertDocumentChunks(ctx, chunks)
}

// objectKey creates a consistent S3 key layout.
func (s *DocumentService) objectKey(userID, docID, filename string) string {
	filename = strings.TrimSpace(filename)
	filename = strings.ReplaceAll(filename, " ", "_")
	return path.Join("users", userID, "documents", docID, filename)
}
