package main

import (
	"log"
	"net/http"
	"time"

	"go-minimal-backend/internal/0config"
	router "go-minimal-backend/internal/1delivery/http"
	"go-minimal-backend/internal/1delivery/http/handler"
	"go-minimal-backend/internal/2usecase"
	"go-minimal-backend/internal/3repository/postgres"
)

func main() {
	// 1. Load configuration
	cfg := config.Load()

	// 2. Setup PostgreSQL
	dbPool, err := postgres.New(cfg.DB_DSN)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbPool.Close()

	// 3. Initialize Repositories
	userRepo := postgres.NewUserRepository(dbPool)
	noteRepo := postgres.NewNoteRepository(dbPool)

	// 4. Initialize Usecases
	tokenExpiration := 24 * time.Hour
	authUseCase := usecase.NewAuthUsecase(userRepo, cfg.JWT_SECRET, tokenExpiration)
	noteUseCase := usecase.NewNoteUsecase(noteRepo)

	// 5. Initialize Handlers
	authHandler := handler.NewAuthHandler(authUseCase)
	noteHandler := handler.NewNoteHandler(noteUseCase)

	// 6. Setup Router
	mux := router.NewRouter(authHandler, noteHandler, cfg.JWT_SECRET, cfg.CORS_ALLOWED_ORIGINS)

	// 7. Start Server
	log.Printf("Starting server on port %s", cfg.PORT)
	server := &http.Server{
		Addr:    ":" + cfg.PORT,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
