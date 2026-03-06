package http

import (
	"net/http"

	"go-minimal-backend/internal/delivery/http/handler"
	"go-minimal-backend/internal/delivery/http/middleware"
	"go-minimal-backend/internal/domain"
)

func NewRouter(authHandler *handler.AuthHandler, noteHandler *handler.NoteHandler, jwtSecret string) *http.ServeMux {
	mux := http.NewServeMux()

	authMw := middleware.Auth(jwtSecret)

	// Public routes
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)

	// Protected routes (Notes CRUD)
	// We wrap the specific handler functions with the auth middleware
	mux.Handle("GET /api/notes", authMw(http.HandlerFunc(noteHandler.List)))
	mux.Handle("POST /api/notes", authMw(http.HandlerFunc(noteHandler.Create)))
	mux.Handle("GET /api/notes/{id}", authMw(http.HandlerFunc(noteHandler.GetByID)))
	mux.Handle("PUT /api/notes/{id}", authMw(http.HandlerFunc(noteHandler.Update)))

	// Example of an Admin-only route
	mux.Handle("DELETE /api/notes/{id}", authMw(middleware.RequireRole(domain.RoleAdmin)(http.HandlerFunc(noteHandler.Delete))))

	return mux
}
