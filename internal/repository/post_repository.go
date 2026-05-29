package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/medivh13/synaptica-api/internal/model"
)

type PostRepository interface {
	UpsertMany(ctx context.Context, posts []model.Post) (PostUpsertResult, error)
	FindByID(ctx context.Context, id string) (*model.Post, error)
	FindRecent(ctx context.Context, filter PostFilter) ([]model.Post, error)
}

type PostUpsertResult struct {
	Inserted int
	Skipped  int
}

type PostFilter struct {
	Source string
	Query  string
	Limit  int
	Offset int
}

type postRepository struct {
	db *sqlx.DB
}

func NewPostRepository(db *sqlx.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) UpsertMany(ctx context.Context, posts []model.Post) (PostUpsertResult, error) {
	if len(posts) == 0 {
		return PostUpsertResult{}, nil
	}

	const query = `
INSERT INTO posts (
    source_id,
    subreddit_id,
    external_id,
    title,
    body,
    author_name,
    score,
    upvote_ratio,
    comment_count,
    permalink,
    url,
    published_at
)
VALUES (
    :source_id,
    :subreddit_id,
    :external_id,
    :title,
    :body,
    :author_name,
    :score,
    :upvote_ratio,
    :comment_count,
    :permalink,
    :url,
    :published_at
)
ON CONFLICT (source_id, external_id)
DO NOTHING;
`

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return PostUpsertResult{}, fmt.Errorf("begin post upsert transaction: %w", err)
	}
	defer tx.Rollback()

	inserted := 0
	for _, post := range posts {
		result, err := tx.NamedExecContext(ctx, query, post)
		if err != nil {
			return PostUpsertResult{}, fmt.Errorf("upsert post %q: %w", post.ExternalID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return PostUpsertResult{}, fmt.Errorf("read post upsert rows affected: %w", err)
		}
		inserted += int(rowsAffected)
	}

	if err := tx.Commit(); err != nil {
		return PostUpsertResult{}, fmt.Errorf("commit post upsert transaction: %w", err)
	}

	return PostUpsertResult{
		Inserted: inserted,
		Skipped:  len(posts) - inserted,
	}, nil
}

func (r *postRepository) FindByID(ctx context.Context, id string) (*model.Post, error) {
	const query = `
SELECT
    id,
    source_id,
    subreddit_id,
    external_id,
    title,
    body,
    author_name,
    score,
    upvote_ratio,
    comment_count,
    permalink,
    url,
    published_at,
    ingested_at,
    updated_at
FROM posts
WHERE id = $1
LIMIT 1;
`

	var post model.Post
	if err := r.db.GetContext(ctx, &post, query, id); err != nil {
		return nil, fmt.Errorf("find post by id %q: %w", id, err)
	}

	return &post, nil
}

func (r *postRepository) FindRecent(ctx context.Context, filter PostFilter) ([]model.Post, error) {
	const query = `
SELECT
    posts.id,
    posts.source_id,
    posts.subreddit_id,
    posts.external_id,
    posts.title,
    posts.body,
    posts.author_name,
    posts.score,
    posts.upvote_ratio,
    posts.comment_count,
    posts.permalink,
    posts.url,
    posts.published_at,
    posts.ingested_at,
    posts.updated_at
FROM posts
JOIN data_sources ON posts.source_id = data_sources.id
WHERE (:source = '' OR data_sources.type = :source)
  AND (
    :query = ''
    OR posts.title ILIKE :search_query
    OR posts.body ILIKE :search_query
  )
ORDER BY posts.published_at DESC
LIMIT :limit
OFFSET :offset;
`

	args := map[string]any{
		"source":       strings.TrimSpace(filter.Source),
		"query":        strings.TrimSpace(filter.Query),
		"search_query": "%" + strings.TrimSpace(filter.Query) + "%",
		"limit":        filter.Limit,
		"offset":       filter.Offset,
	}

	rows, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("find recent posts: %w", err)
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		if err := rows.StructScan(&post); err != nil {
			return nil, fmt.Errorf("scan recent post: %w", err)
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent posts: %w", err)
	}

	return posts, nil
}
