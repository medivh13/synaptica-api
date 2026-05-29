package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"synaptica-api/internal/model"
)

type SubredditRepository interface {
	FindAll(ctx context.Context) ([]model.Subreddit, error)
	FindByName(ctx context.Context, name string) (*model.Subreddit, error)
	FindOrCreateHN(ctx context.Context, sourceID string) (*model.Subreddit, error)
}

type subredditRepository struct {
	db *sqlx.DB
}

func NewSubredditRepository(db *sqlx.DB) SubredditRepository {
	return &subredditRepository{db: db}
}

func (r *subredditRepository) FindAll(ctx context.Context) ([]model.Subreddit, error) {
	const query = `
SELECT
    id,
    source_id,
    name,
    title,
    description,
    subscribers,
    created_at,
    updated_at
FROM subreddits
ORDER BY name ASC;
`

	var subreddits []model.Subreddit
	if err := r.db.SelectContext(ctx, &subreddits, query); err != nil {
		return nil, fmt.Errorf("find all subreddits: %w", err)
	}

	return subreddits, nil
}

func (r *subredditRepository) FindByName(ctx context.Context, name string) (*model.Subreddit, error) {
	const query = `
SELECT
    id,
    source_id,
    name,
    title,
    description,
    subscribers,
    created_at,
    updated_at
FROM subreddits
WHERE LOWER(name) = LOWER($1)
LIMIT 1;
`

	var subreddit model.Subreddit
	if err := r.db.GetContext(ctx, &subreddit, query, name); err != nil {
		return nil, fmt.Errorf("find subreddit by name %q: %w", name, err)
	}

	return &subreddit, nil
}

func (r *subredditRepository) FindOrCreateHN(ctx context.Context, sourceID string) (*model.Subreddit, error) {
	const query = `
INSERT INTO subreddits (
    source_id,
    name,
    title
)
VALUES (
    $1,
    'hackernews',
    'Hacker News'
)
ON CONFLICT (source_id, name)
DO UPDATE SET updated_at = NOW()
RETURNING
    id,
    source_id,
    name,
    title,
    description,
    subscribers,
    created_at,
    updated_at;
`

	var subreddit model.Subreddit
	if err := r.db.GetContext(ctx, &subreddit, query, sourceID); err != nil {
		return nil, fmt.Errorf("find or create hacker news subreddit: %w", err)
	}

	return &subreddit, nil
}
