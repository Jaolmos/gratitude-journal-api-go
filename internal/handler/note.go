package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gratitude-journal-api/internal/model"
	"gratitude-journal-api/internal/service"
)

type NoteHandler struct {
	service *service.NoteService
}

func NewNoteHandler(service *service.NoteService) *NoteHandler {
	return &NoteHandler{service: service}
}

// Estructuras para request/response JSON
type createNoteRequest struct {
	Content string `json:"content"`
	Date    string `json:"date,omitempty"` // formato: 2006-01-02
}

type updateNoteRequest struct {
	Content string `json:"content"`
	Date    string `json:"date,omitempty"`
}

type noteResponse struct {
	ID        int64  `json:"id"`
	Content   string `json:"content"`
	Date      string `json:"date"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// GET /notes
func (h *NoteHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var result []model.Note
	var err error

	if dateStr != "" {
		date, parseErr := time.Parse("2006-01-02", dateStr)
		if parseErr != nil {
			h.respondError(w, http.StatusBadRequest, "formato de fecha inválido, usar YYYY-MM-DD")
			return
		}
		result, err = h.service.GetByDate(date)
	} else if fromStr != "" && toStr != "" {
		from, parseErr := time.Parse("2006-01-02", fromStr)
		if parseErr != nil {
			h.respondError(w, http.StatusBadRequest, "formato de fecha 'from' inválido")
			return
		}
		to, parseErr := time.Parse("2006-01-02", toStr)
		if parseErr != nil {
			h.respondError(w, http.StatusBadRequest, "formato de fecha 'to' inválido")
			return
		}
		result, err = h.service.GetByDateRange(from, to)
	} else {
		result, err = h.service.GetAll()
	}

	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, h.toResponseList(result))
}

// GET /notes/{id}
func (h *NoteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.extractID(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	note, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, service.ErrNoteNotFound) {
			h.respondError(w, http.StatusNotFound, "nota no encontrada")
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, h.toResponse(note))
}

// POST /notes
func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	var date *time.Time
	if req.Date != "" {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "formato de fecha inválido, usar YYYY-MM-DD")
			return
		}
		date = &parsed
	}

	note, err := h.service.Create(req.Content, date)
	if err != nil {
		if errors.Is(err, service.ErrEmptyContent) {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, h.toResponse(note))
}

// PUT /notes/{id}
func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := h.extractID(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req updateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	var date *time.Time
	if req.Date != "" {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "formato de fecha inválido, usar YYYY-MM-DD")
			return
		}
		date = &parsed
	}

	note, err := h.service.Update(id, req.Content, date)
	if err != nil {
		if errors.Is(err, service.ErrNoteNotFound) {
			h.respondError(w, http.StatusNotFound, "nota no encontrada")
			return
		}
		if errors.Is(err, service.ErrEmptyContent) {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, h.toResponse(note))
}

// DELETE /notes/{id}
func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := h.extractID(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.service.Delete(id); err != nil {
		if errors.Is(err, service.ErrNoteNotFound) {
			h.respondError(w, http.StatusNotFound, "nota no encontrada")
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helpers

func (h *NoteHandler) extractID(r *http.Request) (int64, error) {
	path := strings.TrimPrefix(r.URL.Path, "/notes/")
	return strconv.ParseInt(path, 10, 64)
}

func (h *NoteHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *NoteHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, errorResponse{Error: message})
}

func (h *NoteHandler) toResponse(note *model.Note) noteResponse {
	return noteResponse{
		ID:        note.ID,
		Content:   note.Content,
		Date:      note.Date.Format("2006-01-02"),
		CreatedAt: note.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: note.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (h *NoteHandler) toResponseList(notes []model.Note) []noteResponse {
	result := make([]noteResponse, len(notes))
	for i, note := range notes {
		result[i] = noteResponse{
			ID:        note.ID,
			Content:   note.Content,
			Date:      note.Date.Format("2006-01-02"),
			CreatedAt: note.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: note.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result
}
