package service

import (
	"errors"
	"testing"
	"time"

	"gratitude-journal-api/internal/model"
)

// Mock del repository
type mockNoteRepository struct {
	notes    map[int64]*model.Note
	nextID   int64
	findErr  error
	createErr error
	updateErr error
	deleteErr error
}

func newMockRepository() *mockNoteRepository {
	return &mockNoteRepository{
		notes:  make(map[int64]*model.Note),
		nextID: 1,
	}
}

func (m *mockNoteRepository) FindAll() ([]model.Note, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var result []model.Note
	for _, note := range m.notes {
		result = append(result, *note)
	}
	return result, nil
}

func (m *mockNoteRepository) FindByDate(date time.Time) ([]model.Note, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var result []model.Note
	for _, note := range m.notes {
		if note.Date.Format("2006-01-02") == date.Format("2006-01-02") {
			result = append(result, *note)
		}
	}
	return result, nil
}

func (m *mockNoteRepository) FindByDateRange(from, to time.Time) ([]model.Note, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var result []model.Note
	for _, note := range m.notes {
		if !note.Date.Before(from) && !note.Date.After(to) {
			result = append(result, *note)
		}
	}
	return result, nil
}

func (m *mockNoteRepository) FindByID(id int64) (*model.Note, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	note, exists := m.notes[id]
	if !exists {
		return nil, nil
	}
	return note, nil
}

func (m *mockNoteRepository) Create(note *model.Note) error {
	if m.createErr != nil {
		return m.createErr
	}
	note.ID = m.nextID
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()
	m.notes[note.ID] = note
	m.nextID++
	return nil
}

func (m *mockNoteRepository) Update(note *model.Note) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	note.UpdatedAt = time.Now()
	m.notes[note.ID] = note
	return nil
}

func (m *mockNoteRepository) Delete(id int64) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.notes, id)
	return nil
}

func TestCreate_Success(t *testing.T) {
	repo := newMockRepository()
	svc := NewNoteService(repo)

	note, err := svc.Create("Agradecido por el sol", nil)

	if err != nil {
		t.Fatalf("esperaba nil, obtuve error: %v", err)
	}
	if note.ID != 1 {
		t.Errorf("esperaba ID 1, obtuve %d", note.ID)
	}
	if note.Content != "Agradecido por el sol" {
		t.Errorf("contenido incorrecto: %s", note.Content)
	}
}

func TestCreate_WithDate(t *testing.T) {
	repo := newMockRepository()
	svc := NewNoteService(repo)

	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	note, err := svc.Create("Nota con fecha", &date)

	if err != nil {
		t.Fatalf("esperaba nil, obtuve error: %v", err)
	}
	if note.Date.Format("2006-01-02") != "2024-01-15" {
		t.Errorf("fecha incorrecta: %s", note.Date.Format("2006-01-02"))
	}
}

func TestCreate_EmptyContent(t *testing.T) {
	repo := newMockRepository()
	svc := NewNoteService(repo)

	_, err := svc.Create("", nil)

	if !errors.Is(err, ErrEmptyContent) {
		t.Errorf("esperaba ErrEmptyContent, obtuve: %v", err)
	}
}

func TestGetByID_Success(t *testing.T) {
	repo := newMockRepository()
	svc := NewNoteService(repo)

	created, _ := svc.Create("Test nota", nil)
	note, err := svc.GetByID(created.ID)

	if err != nil {
		t.Fatalf("esperaba nil, obtuve error: %v", err)
	}
	if note.Content != "Test nota" {
		t.Errorf("contenido incorrecto: %s", note.Content)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewNoteService(repo)

	_, err := svc.GetByID(999)

	if !errors.Is(err, ErrNoteNotFound) {
		t.Errorf("esperaba ErrNoteNotFound, obtuve: %v", err)
	}
}

func TestUpdate_Success(t *testing.T) {
	repo := newMockRepository()
	svc := NewNoteService(repo)

	created, _ := svc.Create("Original", nil)
	updated, err := svc.Update(created.ID, "Modificado", nil)

	if err != nil {
		t.Fatalf("esperaba nil, obtuve error: %v", err)
	}
	if updated.Content != "Modificado" {
		t.Errorf("contenido incorrecto: %s", updated.Content)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewNoteService(repo)

	_, err := svc.Update(999, "Contenido", nil)

	if !errors.Is(err, ErrNoteNotFound) {
		t.Errorf("esperaba ErrNoteNotFound, obtuve: %v", err)
	}
}

func TestUpdate_EmptyContent(t *testing.T) {
	repo := newMockRepository()
	svc := NewNoteService(repo)

	created, _ := svc.Create("Original", nil)
	_, err := svc.Update(created.ID, "", nil)

	if !errors.Is(err, ErrEmptyContent) {
		t.Errorf("esperaba ErrEmptyContent, obtuve: %v", err)
	}
}

func TestDelete_Success(t *testing.T) {
	repo := newMockRepository()
	svc := NewNoteService(repo)

	created, _ := svc.Create("Para borrar", nil)
	err := svc.Delete(created.ID)

	if err != nil {
		t.Fatalf("esperaba nil, obtuve error: %v", err)
	}

	_, err = svc.GetByID(created.ID)
	if !errors.Is(err, ErrNoteNotFound) {
		t.Errorf("nota no fue eliminada")
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewNoteService(repo)

	err := svc.Delete(999)

	if !errors.Is(err, ErrNoteNotFound) {
		t.Errorf("esperaba ErrNoteNotFound, obtuve: %v", err)
	}
}
