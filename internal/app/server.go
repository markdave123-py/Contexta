package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/markdave123-py/Contexta/internal/api/handlers"
	appMiddleware "github.com/markdave123-py/Contexta/internal/api/middlewares"
	"github.com/markdave123-py/Contexta/internal/config"
	"github.com/markdave123-py/Contexta/internal/core"
	ingestor "github.com/markdave123-py/Contexta/internal/core/ingestion_engine"
)

// // Server wraps the HTTP server instance and its handlers.
// type Server struct {
// 	httpServer *http.Server
// }

// // NewServer builds and wires all routes.
// func NewServer(ctx context.Context, cfg *config.Config, db core.DbClient, obj core.ObjectClient, ing *ingestor.DocumentIngestor, emb core.EmbeddingProvider, llm core.LLMProvider) *Server {
// 	authHandler := handlers.NewAuthHandler(db)
// 	docHandler := handlers.NewDocumentHandler(db, &obj, ing, cfg)
// 	chatHandler := handlers.NewChatHandler(db, emb, llm)

// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.Logger)
// 	r.Use(middleware.Recoverer)
// 	r.Use(middleware.Timeout(60 * time.Second))

// 	r.Use(cors.Handler(cors.Options{
// 		AllowedOrigins:   []string{"http://localhost:5173"},
// 		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
// 		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
// 		AllowCredentials: true,
// 	}))

// 	// public endpoints
// 	r.Post("/api/signup", authHandler.Signup)
// 	r.Post("/api/login", authHandler.Login)

// 	// protected endpoints
// 	r.Group(func(protected chi.Router) {
// 		protected.Use(appMiddleware.JWTMiddleware)
// 		protected.Post("/api/documents/upload", docHandler.UploadDocument)
// 		protected.Get("/api/documents", docHandler.GetDocuments)
// 		protected.Post("/api/chat/query", chatHandler.QueryDocument)
// 	})

// 	httpSrv := &http.Server{
// 		Addr:    ":" + cfg.Port,
// 		Handler: r,
// 	}

// 	return &Server{httpServer: httpSrv}
// }

// // Start runs the HTTP server.
// func (s *Server) Start() {
// 	log.Printf("HTTP server listening on %s", s.httpServer.Addr)
// 	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 		log.Fatalf("server error: %v", err)
// 	}
// }

// // Shutdown gracefully stops the server.
// func (s *Server) Shutdown(ctx context.Context) error {
// 	log.Println("Shutting down HTTP server...")
// 	return s.httpServer.Shutdown(ctx)
// }

// Server wraps the HTTP server instance and its handlers.
type Server struct {
	httpServer *http.Server
}

// NewServer builds and wires all routes.
func NewServer(ctx context.Context, cfg *config.Config, db core.DbClient, obj core.ObjectClient, ing *ingestor.DocumentIngestor, emb core.EmbeddingProvider, llm core.LLMProvider) *Server {
	authHandler := handlers.NewAuthHandler(db)
	docHandler := handlers.NewDocumentHandler(db, &obj, ing, cfg)
	chatHandler := handlers.NewChatHandler(db, emb, llm)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:8888"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// Serve static files from the web directory
	fileServer := http.FileServer(http.Dir("./web"))
	r.Handle("/*", fileServer)

	// API routes
	r.Route("/api", func(api chi.Router) {
		// public endpoints
		api.Post("/signup", authHandler.Signup)
		api.Post("/login", authHandler.Login)

		// protected endpoints
		api.Group(func(protected chi.Router) {
			protected.Use(appMiddleware.JWTMiddleware)
			protected.Post("/documents/upload", docHandler.UploadDocument)
			protected.Get("/documents", docHandler.GetDocuments)
			protected.Post("/chat/query", chatHandler.QueryDocument)

			// Add the chat endpoints your UI expects
			// protected.Post("/documents/{document_id}/chat/ask", chatHandler.QueryDocument)
			// protected.Get("/chat/sessions", chatHandler.GetChatSessions)
			// protected.Get("/chat/sessions/{session_id}", chatHandler.GetChatSession)
		})
	})

	httpSrv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	return &Server{httpServer: httpSrv}
}

// Start runs the HTTP server.
func (s *Server) Start() {
	log.Printf("HTTP server listening on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}
