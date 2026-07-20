package http

import (
	"net/http"

	"go-minimal-backend/internal/1delivery/http/handler"
	"go-minimal-backend/internal/1delivery/http/middleware"
	"go-minimal-backend/internal/4domain"
)

func NewRouter(authHandler *handler.AuthHandler, noteHandler *handler.NoteHandler, jwtSecret string, corsAllowedOrigins []string) http.Handler {
	mux := http.NewServeMux()

	authMw := middleware.Auth(jwtSecret)

	// จำกัด 1 request/วินาทีเฉลี่ยต่อ IP (ยิงรัวได้ 5 ครั้งแรกก่อนโดนบล็อก) เฉพาะ endpoint auth
	// ป้องกัน brute-force เดารหัสผ่าน/สแปมสมัครสมาชิก (ดู ratelimit.go)
	authRateLimit := middleware.RateLimiter(1, 5)

	// Public routes
	mux.Handle("POST /api/auth/register", authRateLimit(http.HandlerFunc(authHandler.Register)))
	mux.Handle("POST /api/auth/login", authRateLimit(http.HandlerFunc(authHandler.Login)))

	// Protected routes (Notes CRUD)
	// We wrap the specific handler functions with the auth middleware

	// กรณีที่ไม่ต้องมีการเช็ค auth middleware
	// mux.HandleFunc("GET /api/notes", noteHandler.List)
	mux.Handle("GET /api/notes", authMw(http.HandlerFunc(noteHandler.List)))
	mux.Handle("POST /api/notes", authMw(http.HandlerFunc(noteHandler.Create)))
	mux.Handle("GET /api/notes/{id}", authMw(http.HandlerFunc(noteHandler.GetByID)))
	mux.Handle("PUT /api/notes/{id}", authMw(http.HandlerFunc(noteHandler.Update)))

	// Example of an Admin-only route
	mux.Handle("DELETE /api/notes/{id}", authMw(middleware.RequireRole(domain.RoleAdmin)(http.HandlerFunc(noteHandler.Delete))))

	// CORS ต้องห่อ mux ทั้งตัวเป็นชั้นนอกสุด เพื่อดัก preflight (OPTIONS) ของทุก route
	// ก่อนที่จะไปถึง mux (ซึ่งไม่มี route ของ OPTIONS ผูกไว้ จะตอบ 404 ถ้าไม่ดักไว้ก่อน)
	wrapped := middleware.CORS(corsAllowedOrigins)(mux)

	// Recover ต้องอยู่นอก CORS แต่ใน RequestID (ดูคอมเมนต์ใน recover.go ว่าทำไมต้องเรียงแบบนี้)
	wrapped = middleware.Recover(wrapped)

	// RequestID ต้องเป็นชั้นนอกสุดจริงๆ เพื่อให้ทุก request ที่เข้ามา (รวมถึง preflight OPTIONS
	// และตอน panic) มี request ID ผูกไว้ตั้งแต่ต้นทาง ก่อนจะไหลผ่าน layer อื่นทั้งหมด
	return middleware.RequestID(wrapped)
}
