package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/markdave123-py/Contexta/internal/app"
	"github.com/markdave123-py/Contexta/internal/config"
)

func main() {
	// Create a background context for the application lifecycle
	ctx := context.Background()

	// Handle SIGINT/SIGTERM for graceful shutdown
	ctx, cancel := context.WithCancel(ctx)

	// Handle SIGINT/SIGTERM for graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		cancel()
	}()

	cfg := config.LoadConfig()
	application, err := app.NewApp(ctx, cfg)

	// application.
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}
	defer application.Close()

	// start the document ingestion worker
	go application.DocProcessor.Start(ctx, cfg.NumProcessors)

	// start HTTP server
	go application.Server.Start()

	log.Println("Contexta server is running; ready to serve request")
	<-ctx.Done()

	cancel()
	application.Server.Shutdown(context.Background())
	log.Println("shutting down...")
}
