package model

import "time"

type Post struct {
	ID           string    `db:"id" json:"id"`
	SourceID     string    `db:"source_id" json:"source_id"`
	SubredditID  string    `db:"subreddit_id" json:"subreddit_id"`
	ExternalID   string    `db:"external_id" json:"external_id"`
	Title        string    `db:"title" json:"title"`
	Body         *string   `db:"body" json:"body,omitempty"`
	AuthorName   *string   `db:"author_name" json:"author_name,omitempty"`
	Score        int       `db:"score" json:"score"`
	UpvoteRatio  *float64  `db:"upvote_ratio" json:"upvote_ratio,omitempty"`
	CommentCount int       `db:"comment_count" json:"comment_count"`
	Permalink    *string   `db:"permalink" json:"permalink,omitempty"`
	URL          *string   `db:"url" json:"url,omitempty"`
	PublishedAt  time.Time `db:"published_at" json:"published_at"`
	IngestedAt   time.Time `db:"ingested_at" json:"ingested_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
