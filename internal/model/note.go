package model

import "time"

type Note struct {
	ID        int64
	Content   string
	Date      time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
