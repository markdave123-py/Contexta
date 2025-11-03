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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT/SIGTERM for graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		cancel()
	}()

	cfg := config.LoadConfig()
	application, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}
	defer application.Close()

	log.Println("Contexta is running; DB connected and bootstrapped.")
	<-ctx.Done()
	log.Println("shutting down...")
}
