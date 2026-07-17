package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"go-minimal-backend/pkg/reqctx"
)

// RequestIDHeader คือชื่อ header ที่ส่ง request ID กลับไปให้ client ทุก response
// (ทั้งตอนสำเร็จและ error) เผื่อ client อยากเก็บไว้เอง หรือส่งกลับมาตอนแจ้งปัญหา
const RequestIDHeader = "X-Request-Id"

// RequestID คือ middleware ตัวแรกสุดที่ request ทุกตัวต้องผ่าน (ห่อไว้ชั้นนอกสุดใน router.go
// ก่อนแม้แต่ CORS) สร้าง ID สุ่มสั้นๆ ต่อ 1 request แล้วผูกไว้ทั้งใน response header และ context
// เพื่อให้ handler/log ทุกจุดของ request นี้อ้างอิง ID เดียวกันได้ ใช้ไล่ log ตอน debug
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := newRequestID()

		w.Header().Set(RequestIDHeader, id)
		ctx := reqctx.WithRequestID(r.Context(), id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// newRequestID สุ่ม 8 byte แล้ว encode เป็น hex (16 ตัวอักษร) สั้นพออ่านง่ายตอน grep log
// ไม่ต้องพึ่ง library uuid เพิ่ม (โปรเจกต์นี้ตั้งใจให้ dependency น้อยที่สุด)
func newRequestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		// แทบไม่เกิดขึ้นจริง (crypto/rand อ่านจาก OS entropy source) แต่กันไว้ไม่ให้ request ค้าง
		return "unknown"
	}
	return hex.EncodeToString(buf)
}
