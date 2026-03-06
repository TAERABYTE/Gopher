package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go-minimal-backend/internal/delivery/http/middleware"
	"go-minimal-backend/internal/domain"
	"go-minimal-backend/pkg/response"
)

type NoteHandler struct {
	noteUsecase domain.NoteUsecase
}

func NewNoteHandler(nu domain.NoteUsecase) *NoteHandler {
	return &NoteHandler{noteUsecase: nu}
}

type createNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req createNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	note := &domain.Note{
		UserID:  user.ID,
		Title:   req.Title,
		Content: req.Content,
	}

	if err := h.noteUsecase.Create(r.Context(), note); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create note")
		return
	}

	response.JSON(w, http.StatusCreated, note)
}

func (h *NoteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	note, err := h.noteUsecase.GetByID(r.Context(), id, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Note not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to get note")
		return
	}

	response.JSON(w, http.StatusOK, note)
}

func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	notes, err := h.noteUsecase.List(r.Context(), user.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list notes")
		return
	}

	// return empty array instead of null
	if notes == nil {
		notes = []*domain.Note{}
	}

	response.JSON(w, http.StatusOK, notes)
}

func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	var req createNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	note := &domain.Note{
		ID:      id,
		UserID:  user.ID,
		Title:   req.Title,
		Content: req.Content,
	}

	if err := h.noteUsecase.Update(r.Context(), note); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Note not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to update note")
		return
	}

	// fetch updated note to return it
	updatedNote, _ := h.noteUsecase.GetByID(r.Context(), id, user.ID)
	response.JSON(w, http.StatusOK, updatedNote)
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	if err := h.noteUsecase.Delete(r.Context(), id, user.ID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Note not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to delete note")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Note deleted successfully"})
}
