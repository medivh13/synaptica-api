package model

import "time"

type Comment struct {
	ID               string    `db:"id" json:"id"`
	SourceID         string    `db:"source_id" json:"source_id"`
	SubredditID      string    `db:"subreddit_id" json:"subreddit_id"`
	PostID           string    `db:"post_id" json:"post_id"`
	ExternalID       string    `db:"external_id" json:"external_id"`
	ParentExternalID *string   `db:"parent_external_id" json:"parent_external_id,omitempty"`
	ParentCommentID  *string   `db:"parent_comment_id" json:"parent_comment_id,omitempty"`
	AuthorName       *string   `db:"author_name" json:"author_name,omitempty"`
	Body             string    `db:"body" json:"body"`
	Score            int       `db:"score" json:"score"`
	Depth            int       `db:"depth" json:"depth"`
	PublishedAt      time.Time `db:"published_at" json:"published_at"`
	IngestedAt       time.Time `db:"ingested_at" json:"ingested_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}
