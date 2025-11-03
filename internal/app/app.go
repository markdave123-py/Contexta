// internal/app/app.go
package app

import (
	"context"
	"log"

	"github.com/markdave123-py/Contexta/internal/config"
	db "github.com/markdave123-py/Contexta/internal/core/database"
	objectclient "github.com/markdave123-py/Contexta/internal/core/object-client"
)

type App struct {
	DBClient     *db.DatabaseClient
	ObjectClient *objectclient.S3Client
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	dbClient, err := db.NewDatabaseClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	log.Println("Database initialized and ready.")

	obClient, err := objectclient.NewS3Client(ctx, cfg)

	if err != nil{
		return nil, err
	}

	log.Println("Object client initialized and ready.")
	
	return &App{DBClient: dbClient.(*db.DatabaseClient), ObjectClient: obClient.(*objectclient.S3Client)}, nil
}

func (a *App) Close() {
	if a.DBClient != nil {
		_ = a.DBClient.Close()
	}
}
