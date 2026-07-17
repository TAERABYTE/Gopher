package middleware

import "net/http"

// CORS สร้าง middleware ที่ห่อ handler ทั้งหมดของแอป (ใช้ครอบ mux ทั้งตัวใน router.go
// ไม่ใช่ครอบทีละ route แบบ Auth) หน้าที่หลักคือใส่ header Access-Control-Allow-*
// ให้ browser ฝั่ง frontend (คนละ origin/port) อนุญาตให้ยิง request เข้ามาได้
//
// allowedOrigins: list ของ origin ที่อนุญาต (เช่น "http://localhost:3000")
// ถ้าใส่ "*" ตัวเดียวในลิสต์ = อนุญาตทุก origin (เหมาะกับตอน dev เท่านั้น)
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" {
				if allowAll {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if originAllowed(origin, allowedOrigins) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					// Vary: Origin บอก cache/proxy ว่า response นี้ขึ้นอยู่กับ Origin header
					// ป้องกัน response ของ origin นึงถูก cache แล้วส่งให้อีก origin นึงผิดๆ
					w.Header().Set("Vary", "Origin")
					// ใส่ credentials ได้เฉพาะตอนระบุ origin ชัดเจนเท่านั้น (spec ห้ามคู่กับ "*")
					// ตอนนี้ยังไม่ได้ใช้ cookie (auth เป็น JWT ผ่าน Authorization header)
					// แต่เผื่อไว้ถ้าอนาคตเปลี่ยนไปใช้ cookie-based session/refresh token
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Preflight request: browser ยิง OPTIONS มาถามก่อนว่ายิงจริงได้ไหม
			// (เกิดตอนมี custom header อย่าง Authorization หรือ method อย่าง PUT/DELETE)
			// ต้องตอบจบตรงนี้เลย ห้ามส่งต่อไปที่ mux เพราะ mux ไม่ได้ผูก route ของ OPTIONS ไว้
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func originAllowed(origin string, allowed []string) bool {
	for _, o := range allowed {
		if o == origin {
			return true
		}
	}
	return false
}
