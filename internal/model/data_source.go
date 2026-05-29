package model

import "time"

type DataSource struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Type      string    `db:"type" json:"type"`
	BaseURL   *string   `db:"base_url" json:"base_url,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
