package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"go-minimal-backend/pkg/response"

	"golang.org/x/time/rate"
)

// visitor เก็บ token bucket ของ IP หนึ่งตัว พร้อมเวลาที่ใช้ล่าสุด (ไว้ทำ cleanup ทิ้งของเก่า)
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter จำกัดจำนวน request ต่อ IP หนึ่งตัว ใช้ครอบเฉพาะ endpoint ที่เสี่ยงถูก brute-force
// เช่น /api/auth/login, /api/auth/register (ดู router.go) ไม่ได้ครอบทั้ง API เพราะ endpoint
// ทั่วไป (CRUD notes หลัง login แล้ว) ไม่ได้เสี่ยงแบบเดียวกัน และ user ปกติก็ยิงถี่ได้อยู่แล้ว
//
// rps  = จำนวน request เฉลี่ยที่อนุญาตต่อวินาทีต่อ IP (เติม token คืนอัตรานี้)
// burst = จำนวน request ที่ยิงรัวๆ ติดกันได้ในทีเดียวก่อนโดนบล็อก (ขนาดถังเก็บ token)
func RateLimiter(rps float64, burst int) func(http.Handler) http.Handler {
	var (
		mu       sync.Mutex
		visitors = make(map[string]*visitor)
	)

	// cleanup goroutine ทิ้ง IP ที่เงียบไปนานเกิน 3 นาที กัน map โตไม่มีที่สิ้นสุดถ้ามี IP
	// แปลกใหม่เข้ามาเรื่อยๆ (เช่นโดน scan) รันตลอดอายุของ process เพราะ middleware นี้สร้างครั้งเดียวตอน boot
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()

		v, exists := visitors[ip]
		if !exists {
			limiter := rate.NewLimiter(rate.Limit(rps), burst)
			visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
			return limiter
		}
		v.lastSeen = time.Now()
		return v.limiter
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if !getLimiter(ip).Allow() {
				response.Error(w, r, http.StatusTooManyRequests, "Too many requests, please try again later")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP พยายามอ่าน IP จริงของ client จาก X-Forwarded-For ก่อน (กรณีอยู่หลัง reverse proxy)
// ถ้าไม่มีค่อย fallback ไปใช้ r.RemoteAddr ตรงๆ
//
// ข้อควรระวัง: X-Forwarded-For ปลอมแปลงได้ถ้าไม่มี proxy ที่เชื่อถือได้คอยเซ็ตให้เสมอ
// ตอน deploy จริงหลัง reverse proxy ต้องตั้งค่า proxy ให้ overwrite header นี้เอง ห้ามปล่อยให้
// client ส่งมาเองตรงๆ ไม่งั้น attacker ปลอม header นี้เพื่อหลบ rate limit ได้
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
