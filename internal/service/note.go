package service

import (
	"errors"
	"time"

	"gratitude-journal-api/internal/model"
)

var (
	ErrEmptyContent = errors.New("el contenido no puede estar vacío")
	ErrNoteNotFound = errors.New("nota no encontrada")
)

// Interface definida donde se consume (principio de Go)
type NoteRepository interface {
	FindAll() ([]model.Note, error)
	FindByDate(date time.Time) ([]model.Note, error)
	FindByDateRange(from, to time.Time) ([]model.Note, error)
	FindByID(id int64) (*model.Note, error)
	Create(note *model.Note) error
	Update(note *model.Note) error
	Delete(id int64) error
}

type NoteService struct {
	repo NoteRepository
}

func NewNoteService(repo NoteRepository) *NoteService {
	return &NoteService{repo: repo}
}

func (s *NoteService) GetAll() ([]model.Note, error) {
	return s.repo.FindAll()
}

func (s *NoteService) GetByDate(date time.Time) ([]model.Note, error) {
	return s.repo.FindByDate(date)
}

func (s *NoteService) GetByDateRange(from, to time.Time) ([]model.Note, error) {
	return s.repo.FindByDateRange(from, to)
}

func (s *NoteService) GetByID(id int64) (*model.Note, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, ErrNoteNotFound
	}
	return note, nil
}

func (s *NoteService) Create(content string, date *time.Time) (*model.Note, error) {
	if content == "" {
		return nil, ErrEmptyContent
	}

	note := &model.Note{
		Content: content,
	}

	// Si no se especifica fecha, usar la fecha actual
	if date != nil {
		note.Date = *date
	} else {
		note.Date = time.Now().Truncate(24 * time.Hour)
	}

	if err := s.repo.Create(note); err != nil {
		return nil, err
	}

	return note, nil
}

func (s *NoteService) Update(id int64, content string, date *time.Time) (*model.Note, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, ErrNoteNotFound
	}

	if content == "" {
		return nil, ErrEmptyContent
	}

	note.Content = content
	if date != nil {
		note.Date = *date
	}

	if err := s.repo.Update(note); err != nil {
		return nil, err
	}

	return note, nil
}

func (s *NoteService) Delete(id int64) error {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if note == nil {
		return ErrNoteNotFound
	}

	return s.repo.Delete(id)
}
