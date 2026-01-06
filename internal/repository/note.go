package repository

import (
	"database/sql"
	"time"

	"gratitude-journal-api/internal/model"
)

const timeFormat = "2006-01-02 15:04:05"
const dateFormat = "2006-01-02"

type NoteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) FindAll() ([]model.Note, error) {
	query := `SELECT id, content, date, created_at, updated_at FROM notes ORDER BY date DESC, id DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

func (r *NoteRepository) FindByDate(date time.Time) ([]model.Note, error) {
	query := `SELECT id, content, date, created_at, updated_at FROM notes WHERE date = ? ORDER BY id DESC`

	rows, err := r.db.Query(query, date.Format(dateFormat))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

func (r *NoteRepository) FindByDateRange(from, to time.Time) ([]model.Note, error) {
	query := `SELECT id, content, date, created_at, updated_at FROM notes WHERE date >= ? AND date <= ? ORDER BY date DESC, id DESC`

	rows, err := r.db.Query(query, from.Format(dateFormat), to.Format(dateFormat))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

func (r *NoteRepository) FindByID(id int64) (*model.Note, error) {
	query := `SELECT id, content, date, created_at, updated_at FROM notes WHERE id = ?`

	row := r.db.QueryRow(query, id)

	var note model.Note
	var dateStr, createdStr, updatedStr string

	err := row.Scan(&note.ID, &note.Content, &dateStr, &createdStr, &updatedStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	note.Date, _ = time.Parse(dateFormat, dateStr)
	note.CreatedAt, _ = time.Parse(timeFormat, createdStr)
	note.UpdatedAt, _ = time.Parse(timeFormat, updatedStr)

	return &note, nil
}

func (r *NoteRepository) Create(note *model.Note) error {
	query := `INSERT INTO notes (content, date, created_at, updated_at) VALUES (?, ?, ?, ?)`

	now := time.Now()
	note.CreatedAt = now
	note.UpdatedAt = now

	result, err := r.db.Exec(query,
		note.Content,
		note.Date.Format(dateFormat),
		note.CreatedAt.Format(timeFormat),
		note.UpdatedAt.Format(timeFormat),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	note.ID = id
	return nil
}

func (r *NoteRepository) Update(note *model.Note) error {
	query := `UPDATE notes SET content = ?, date = ?, updated_at = ? WHERE id = ?`

	note.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		note.Content,
		note.Date.Format(dateFormat),
		note.UpdatedAt.Format(timeFormat),
		note.ID,
	)
	return err
}

func (r *NoteRepository) Delete(id int64) error {
	query := `DELETE FROM notes WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *NoteRepository) scanNotes(rows *sql.Rows) ([]model.Note, error) {
	var notes []model.Note

	for rows.Next() {
		var note model.Note
		var dateStr, createdStr, updatedStr string

		err := rows.Scan(&note.ID, &note.Content, &dateStr, &createdStr, &updatedStr)
		if err != nil {
			return nil, err
		}

		note.Date, _ = time.Parse(dateFormat, dateStr)
		note.CreatedAt, _ = time.Parse(timeFormat, createdStr)
		note.UpdatedAt, _ = time.Parse(timeFormat, updatedStr)

		notes = append(notes, note)
	}

	return notes, rows.Err()
}
