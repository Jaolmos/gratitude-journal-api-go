package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gratitude-journal-api/internal/model"
	"gratitude-journal-api/internal/service"
)

// Mock del service
type mockNoteService struct {
	notes   map[int64]*model.Note
	nextID  int64
}

func newMockService() *mockNoteService {
	return &mockNoteService{
		notes:  make(map[int64]*model.Note),
		nextID: 1,
	}
}

func (m *mockNoteService) GetAll() ([]model.Note, error) {
	var result []model.Note
	for _, note := range m.notes {
		result = append(result, *note)
	}
	return result, nil
}

func (m *mockNoteService) GetByDate(date time.Time) ([]model.Note, error) {
	var result []model.Note
	for _, note := range m.notes {
		if note.Date.Format("2006-01-02") == date.Format("2006-01-02") {
			result = append(result, *note)
		}
	}
	return result, nil
}

func (m *mockNoteService) GetByDateRange(from, to time.Time) ([]model.Note, error) {
	var result []model.Note
	for _, note := range m.notes {
		if !note.Date.Before(from) && !note.Date.After(to) {
			result = append(result, *note)
		}
	}
	return result, nil
}

func (m *mockNoteService) GetByID(id int64) (*model.Note, error) {
	note, exists := m.notes[id]
	if !exists {
		return nil, service.ErrNoteNotFound
	}
	return note, nil
}

func (m *mockNoteService) Create(content string, date *time.Time) (*model.Note, error) {
	if content == "" {
		return nil, service.ErrEmptyContent
	}

	note := &model.Note{
		ID:        m.nextID,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if date != nil {
		note.Date = *date
	} else {
		note.Date = time.Now().Truncate(24 * time.Hour)
	}

	m.notes[note.ID] = note
	m.nextID++
	return note, nil
}

func (m *mockNoteService) Update(id int64, content string, date *time.Time) (*model.Note, error) {
	note, exists := m.notes[id]
	if !exists {
		return nil, service.ErrNoteNotFound
	}
	if content == "" {
		return nil, service.ErrEmptyContent
	}

	note.Content = content
	if date != nil {
		note.Date = *date
	}
	note.UpdatedAt = time.Now()
	return note, nil
}

func (m *mockNoteService) Delete(id int64) error {
	if _, exists := m.notes[id]; !exists {
		return service.ErrNoteNotFound
	}
	delete(m.notes, id)
	return nil
}

func TestHandler_Create_Success(t *testing.T) {
	svc := newMockService()
	h := NewNoteHandler(svc)

	body := `{"content": "Agradecido por este test"}`
	req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusCreated, rec.Code)
	}

	var resp noteResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Content != "Agradecido por este test" {
		t.Errorf("contenido incorrecto: %s", resp.Content)
	}
}

func TestHandler_Create_EmptyContent(t *testing.T) {
	svc := newMockService()
	h := NewNoteHandler(svc)

	body := `{"content": ""}`
	req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	svc := newMockService()
	h := NewNoteHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString("invalid"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandler_GetByID_Success(t *testing.T) {
	svc := newMockService()
	h := NewNoteHandler(svc)

	svc.Create("Test nota", nil)

	req := httptest.NewRequest(http.MethodGet, "/notes/1", nil)
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusOK, rec.Code)
	}

	var resp noteResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ID != 1 {
		t.Errorf("esperaba ID 1, obtuve %d", resp.ID)
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	svc := newMockService()
	h := NewNoteHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/notes/999", nil)
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusNotFound, rec.Code)
	}
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	svc := newMockService()
	h := NewNoteHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/notes/abc", nil)
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	svc := newMockService()
	h := NewNoteHandler(svc)

	svc.Create("Original", nil)

	body := `{"content": "Modificado"}`
	req := httptest.NewRequest(http.MethodPut, "/notes/1", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusOK, rec.Code)
	}

	var resp noteResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Content != "Modificado" {
		t.Errorf("contenido incorrecto: %s", resp.Content)
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	svc := newMockService()
	h := NewNoteHandler(svc)

	body := `{"content": "Contenido"}`
	req := httptest.NewRequest(http.MethodPut, "/notes/999", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusNotFound, rec.Code)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	svc := newMockService()
	h := NewNoteHandler(svc)

	svc.Create("Para borrar", nil)

	req := httptest.NewRequest(http.MethodDelete, "/notes/1", nil)
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusNoContent, rec.Code)
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	svc := newMockService()
	h := NewNoteHandler(svc)

	req := httptest.NewRequest(http.MethodDelete, "/notes/999", nil)
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusNotFound, rec.Code)
	}
}

func TestHandler_GetAll_Success(t *testing.T) {
	svc := newMockService()
	h := NewNoteHandler(svc)

	svc.Create("Nota 1", nil)
	svc.Create("Nota 2", nil)

	req := httptest.NewRequest(http.MethodGet, "/notes", nil)
	rec := httptest.NewRecorder()

	h.GetAll(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusOK, rec.Code)
	}

	var resp []noteResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Errorf("esperaba 2 notas, obtuve %d", len(resp))
	}
}
