package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/markdave123-py/Contexta/internal/config"
	db "github.com/markdave123-py/Contexta/internal/core/database"
	"github.com/markdave123-py/Contexta/internal/core/ingestion_engine"
	objectclient "github.com/markdave123-py/Contexta/internal/core/object-client"
	"github.com/markdave123-py/Contexta/internal/models"
)

type DocumentHandler struct {
	dbclient     db.DbClient
	objectclient objectclient.ObjectClient
	ingestor     ingestion_engine.Ingestor
	cfg          *config.Config
}

func NewDocumentHandler(dbclient db.DbClient, objectclient objectclient.ObjectClient, ing ingestion_engine.Ingestor, cfg *config.Config) *DocumentHandler {
	return &DocumentHandler{dbclient: dbclient, objectclient: objectclient, ingestor: ing, cfg: cfg}
}

// UploadDocument handles file upload, DB insert, and background processing.
func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(52 << 20) // 352 MB

	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "user_id not found in context", http.StatusUnauthorized)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		println(err.Error())
		http.Error(w, "invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// --- Key Generation for S3 ---
	// Sanitize filename to prevent path traversal or invalid characters
	cleanFilename := filepath.Base(header.Filename) // Removes any path components
	docID := uuid.NewString()

	s3Key := fmt.Sprintf("%s/%s/%s", userID, docID, cleanFilename)

	// Get Content-Type from header
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream" // Default if not provided
	}

	uploadctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	url, err := h.objectclient.UploadFile(uploadctx, h.cfg.BucketName, s3Key, file, contentType)
	if err != nil {
		http.Error(w, fmt.Sprintf("upload failed: %v", err), 500)
		return
	}

	doc := &models.Document{
		ID:          uuid.NewString(),
		UserID:      userID,
		FileName:    header.Filename,
		StorageURL:  url,
		SourceType:  "upload",
		Status:      "uploaded",
		ContentType: contentType,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.dbclient.CreateDocument(uploadctx, doc); err != nil {
		log.Printf("DB insert failed for doc %s: %v", docID, err)
		http.Error(w, fmt.Sprintf("failed to store document metadata: %v", err), http.StatusInternalServerError)
		return
	}

	h.ingestor.Enqueue(doc.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func (h *DocumentHandler) GetDocuments(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "user_id not found in context", http.StatusUnauthorized)
		return
	}

	documents, err := h.dbclient.ListDocumentsByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}
