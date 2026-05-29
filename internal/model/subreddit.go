package model

import "time"

type Subreddit struct {
	ID          string    `db:"id" json:"id"`
	SourceID    string    `db:"source_id" json:"source_id"`
	Name        string    `db:"name" json:"name"`
	Title       *string   `db:"title" json:"title,omitempty"`
	Description *string   `db:"description" json:"description,omitempty"`
	Subscribers *int64    `db:"subscribers" json:"subscribers,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
