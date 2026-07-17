package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"go-minimal-backend/internal/4domain"
	"go-minimal-backend/pkg/reqctx"
	"go-minimal-backend/pkg/response"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

func NewAuthHandler(au domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: au}
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Username == "" || req.Password == "" {
		response.Error(w, r, http.StatusBadRequest, "Username and password are required")
		return
	}

	err := h.authUsecase.Register(r.Context(), req.Username, req.Password, domain.RoleUser)
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			response.Error(w, r, http.StatusConflict, "Username already exists")
			return
		}
		// log รายละเอียด error จริงไว้ฝั่ง server เท่านั้น (อาจมี SQL/connection detail ปนอยู่)
		// ส่วน client ได้รับแค่ข้อความกลางๆ ไม่ให้หลุดข้อมูล internal ออกไป
		log.Printf("[%s] auth: register failed: %v", reqctx.RequestID(r.Context()), err)
		response.Error(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, token, err := h.authUsecase.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCreds) {
			response.Error(w, r, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		log.Printf("[%s] auth: login failed: %v", reqctx.RequestID(r.Context()), err)
		response.Error(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"username": user.Username, "token": token, })
}
