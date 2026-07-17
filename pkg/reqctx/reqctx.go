// Package reqctx เก็บ request ID ไว้ใน context ของแต่ละ request
// แยกเป็น package เล็กๆ ต่างหาก (ไม่รวมไว้ใน middleware) เพราะทั้ง middleware (คนเซ็ตค่า)
// และ pkg/response (คนอ่านค่าไปใส่ใน error body) ต้อง import อันนี้ร่วมกัน
// ถ้าให้ response import middleware ตรงๆ จะเกิด import cycle (middleware ก็ import response อยู่แล้ว)
package reqctx

import "context"

type contextKey string

const requestIDKey = contextKey("request_id")

// WithRequestID คืน context ใหม่ที่ผูก request ID ไว้
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID ดึง request ID ออกจาก context (คืนค่าว่างถ้าไม่มี เช่น เรียกนอก request จริง)
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}
