package model

import "time"

type Keyword struct {
	ID          string    `db:"id" json:"id"`
	Keyword     string    `db:"keyword" json:"keyword"`
	Description *string   `db:"description" json:"description,omitempty"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
