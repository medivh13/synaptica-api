package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"synaptica-api/internal/model"
)

type KeywordRepository interface {
	Create(ctx context.Context, input KeywordCreateInput) (*model.Keyword, error)
	FindAllActive(ctx context.Context) ([]model.Keyword, error)
	FindByID(ctx context.Context, id string) (*model.Keyword, error)
	MatchPosts(ctx context.Context, keyword model.Keyword) (KeywordMatchResult, error)
	FindPosts(ctx context.Context, keywordID string, limit int, offset int) ([]model.Post, error)
}

type KeywordCreateInput struct {
	Keyword     string
	Description *string
}

type KeywordMatchResult struct {
	Matched  int
	Inserted int
	Skipped  int
}

type keywordRepository struct {
	db *sqlx.DB
}

func NewKeywordRepository(db *sqlx.DB) KeywordRepository {
	return &keywordRepository{db: db}
}

func (r *keywordRepository) Create(ctx context.Context, input KeywordCreateInput) (*model.Keyword, error) {
	const query = `
WITH inserted AS (
    INSERT INTO keywords (
        keyword,
        description
    )
    VALUES (
        $1,
        $2
    )
    ON CONFLICT (keyword) DO NOTHING
    RETURNING id, keyword, description, is_active, created_at
)
SELECT id, keyword, description, is_active, created_at
FROM inserted
UNION ALL
SELECT id, keyword, description, is_active, created_at
FROM keywords
WHERE keyword = $1
LIMIT 1;
`

	var keyword model.Keyword
	if err := r.db.GetContext(ctx, &keyword, query, input.Keyword, input.Description); err != nil {
		return nil, fmt.Errorf("create keyword %q: %w", input.Keyword, err)
	}

	return &keyword, nil
}

func (r *keywordRepository) FindAllActive(ctx context.Context) ([]model.Keyword, error) {
	const query = `
SELECT id, keyword, description, is_active, created_at
FROM keywords
WHERE is_active = TRUE
ORDER BY keyword ASC;
`

	var keywords []model.Keyword
	if err := r.db.SelectContext(ctx, &keywords, query); err != nil {
		return nil, fmt.Errorf("find active keywords: %w", err)
	}

	return keywords, nil
}

func (r *keywordRepository) FindByID(ctx context.Context, id string) (*model.Keyword, error) {
	const query = `
SELECT id, keyword, description, is_active, created_at
FROM keywords
WHERE id = $1
LIMIT 1;
`

	var keyword model.Keyword
	if err := r.db.GetContext(ctx, &keyword, query, id); err != nil {
		return nil, fmt.Errorf("find keyword by id %q: %w", id, err)
	}

	return &keyword, nil
}

func (r *keywordRepository) MatchPosts(ctx context.Context, keyword model.Keyword) (KeywordMatchResult, error) {
	const countQuery = `
SELECT COUNT(*)
FROM posts
WHERE title ILIKE $1
   OR body ILIKE $1;
`

	const insertQuery = `
INSERT INTO keyword_matches (
    keyword_id,
    entity_type,
    entity_id,
    matched_text
)
SELECT
    $1,
    'post',
    posts.id,
    posts.title
FROM posts
WHERE posts.title ILIKE $2
   OR posts.body ILIKE $2
ON CONFLICT (keyword_id, entity_type, entity_id)
DO NOTHING;
`

	searchQuery := "%" + strings.TrimSpace(keyword.Keyword) + "%"

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return KeywordMatchResult{}, fmt.Errorf("begin keyword post matching transaction: %w", err)
	}
	defer tx.Rollback()

	var matched int
	if err := tx.GetContext(ctx, &matched, countQuery, searchQuery); err != nil {
		return KeywordMatchResult{}, fmt.Errorf("count keyword post matches: %w", err)
	}

	result, err := tx.ExecContext(ctx, insertQuery, keyword.ID, searchQuery)
	if err != nil {
		return KeywordMatchResult{}, fmt.Errorf("insert keyword post matches: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return KeywordMatchResult{}, fmt.Errorf("read keyword post match rows affected: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return KeywordMatchResult{}, fmt.Errorf("commit keyword post matching transaction: %w", err)
	}

	inserted := int(rowsAffected)
	return KeywordMatchResult{
		Matched:  matched,
		Inserted: inserted,
		Skipped:  matched - inserted,
	}, nil
}

func (r *keywordRepository) FindPosts(ctx context.Context, keywordID string, limit int, offset int) ([]model.Post, error) {
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
JOIN keyword_matches
  ON keyword_matches.entity_id = posts.id
 AND keyword_matches.entity_type = 'post'
WHERE keyword_matches.keyword_id = $1
ORDER BY posts.published_at DESC
LIMIT $2
OFFSET $3;
`

	var posts []model.Post
	if err := r.db.SelectContext(ctx, &posts, query, keywordID, limit, offset); err != nil {
		return nil, fmt.Errorf("find keyword posts: %w", err)
	}

	return posts, nil
}
