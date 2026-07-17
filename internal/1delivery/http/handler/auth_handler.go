package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-minimal-backend/internal/4domain"
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
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Username == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	err := h.authUsecase.Register(r.Context(), req.Username, req.Password, domain.RoleUser)
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			response.Error(w, http.StatusConflict, "Username already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, token, err := h.authUsecase.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCreds) {
			response.Error(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"username": user.Username, "token": token, })
}
