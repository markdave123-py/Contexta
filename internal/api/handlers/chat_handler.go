package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/markdave123-py/Contexta/internal/core"
	db "github.com/markdave123-py/Contexta/internal/core/database"
)

type ChatHandler struct {
	dbclient db.DbClient
	embedder core.EmbeddingProvider
	llm      core.LLMProvider
}

func NewChatHandler(db db.DbClient, emb core.EmbeddingProvider, llm core.LLMProvider) *ChatHandler {
	return &ChatHandler{dbclient: db, embedder: emb, llm: llm}
}

type ChatRequest struct {
	DocumentID string `json:"document_id"`
	Query      string `json:"query"`
}

func (h *ChatHandler) QueryDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}

	// Confirm document belongs to user
	doc, err := h.dbclient.GetDocumentByID(ctx, req.DocumentID)
	if err != nil || doc == nil {
		http.Error(w, "document not found", http.StatusNotFound)
		return
	}

	if doc.UserID != userID {
		http.Error(w, "you are unauthoriazed to access this document", http.StatusUnauthorized)
	}

	// Embed the query
	vecs, err := h.embedder.EmbedTexts(ctx, []string{req.Query})
	if err != nil || len(vecs) == 0 {
		http.Error(w, fmt.Sprintf("embedding failed: %v", err), 500)
		return
	}
	queryVec := vecs[0]

	// Retrieve top chunks
	chunks, err := h.dbclient.SearchDocumentChunks(ctx, req.DocumentID, queryVec, 5)
	if err != nil {
		http.Error(w, fmt.Sprintf("search failed: %v", err), 500)
		return
	}

	// 3️⃣ Build context prompt
	var sb strings.Builder
	for _, ch := range chunks {
		sb.WriteString(ch.Text)
		sb.WriteString("\n---\n")
	}

	systemPrompt := "You are an intelligent assistant answering based only on the given document content. If unsure, say 'I cannot find this in the document.'"
	userPrompt := fmt.Sprintf("Context:\n%s\n\nQuestion: %s", sb.String(), req.Query)

	// Generate response
	answer, err := h.llm.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		http.Error(w, fmt.Sprintf("LLM failed: %v", err), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"answer": answer,
	})
}
