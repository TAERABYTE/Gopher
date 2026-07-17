package response

import (
	"encoding/json"
	"net/http"

	"go-minimal-backend/pkg/reqctx"
)

type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// Error ตอบ error กลับไปพร้อม request_id ของ request นี้ (ดึงมาจาก context ที่
// middleware.RequestID ผูกไว้ตั้งแต่ต้นทาง) client เอา request_id นี้ไปแจ้งปัญหาได้
// แล้ว dev grep หา error จริงใน server log ที่ log ไว้คู่กับ ID เดียวกันเจอทันที
func Error(w http.ResponseWriter, r *http.Request, status int, err string) {
	JSON(w, status, ErrorResponse{Error: err, RequestID: reqctx.RequestID(r.Context())})
}

// ValidationErrors ตอบ 400 พร้อม error รายฟิลด์ เช่น {"errors": {"title": "title is required"}}
// ใช้ตอน request ผ่านการ decode JSON ได้ แต่ค่าที่ส่งมาไม่ผ่านกฎ validation (เช่น field required, ความยาวเกิน)
// รูปแบบ per-field แบบนี้ทำให้ frontend เอาไปโชว์ error ใต้ input แต่ละช่องของฟอร์มได้ตรงๆ
func ValidationErrors(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	JSON(w, http.StatusBadRequest, map[string]any{
		"errors":     errors,
		"request_id": reqctx.RequestID(r.Context()),
	})
}
