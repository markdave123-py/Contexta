// internal/app/app.go
package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/markdave123-py/Contexta/internal/config"
	db "github.com/markdave123-py/Contexta/internal/core/database"
	"github.com/markdave123-py/Contexta/internal/core/ingestion_engine"
	"github.com/markdave123-py/Contexta/internal/core/llm"
	objectclient "github.com/markdave123-py/Contexta/internal/core/object-client"
)

type App struct {
	DBClient     db.DbClient
	ObjectClient objectclient.ObjectClient
	DocProcessor ingestion_engine.Ingestor
	Server       *Server
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	appCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	dbClient, err := db.NewDatabaseClient(appCtx, cfg)
	if err != nil {
		return nil, err
	}
	log.Println("Database initialized and ready.")

	objClient, err := objectclient.NewS3Client(appCtx, cfg)

	if err != nil {
		return nil, err
	}
	log.Println("Object client initialized and ready.")

	geminiEmbedder, err := llm.NewGeminiEmbedder(appCtx, cfg.AIAPIKey, cfg.EmbedModel)
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize the embedder, %w", err)
	}

	llmProvider, err := llm.NewGeminiLLM(appCtx, cfg.AIAPIKey, cfg.GenModel)

	if err != nil {
		return nil, fmt.Errorf("couldn't initialize the embedder, %w", err)
	}

	useReadability := false
	documentExtractor := ingestion_engine.NewDocconvExtractor(useReadability)

	ingCfg := &ingestion_engine.IngestConfig{
		TargetTokens:  100,
		OverlapTokens: 5,
		BatchSize:     16,
	}

	docIngestor := ingestion_engine.NewDocumentIngestor(dbClient, objClient, geminiEmbedder, documentExtractor, ingCfg)

	server := NewServer(context.Background(), cfg, dbClient, objClient, docIngestor, geminiEmbedder, llmProvider)

	return &App{DBClient: dbClient.(*db.DatabaseClient), ObjectClient: objClient.(*objectclient.S3Client), DocProcessor: docIngestor, Server: server}, nil
}

func (a *App) Close() {
	if a.DBClient != nil {
		_ = a.DBClient.Close()
	}
}
