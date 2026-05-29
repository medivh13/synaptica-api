package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/medivh13/synaptica-api/internal/model"
)

type CommentRepository interface {
	UpsertMany(ctx context.Context, comments []model.Comment) (CommentUpsertResult, error)
	FindByPostID(ctx context.Context, filter CommentFilter) ([]model.Comment, error)
}

type CommentFilter struct {
	PostID string
	Limit  int
	Offset int
	Depth  *int
}

type CommentUpsertResult struct {
	Inserted int
	Skipped  int
}

type commentRepository struct {
	db *sqlx.DB
}

func NewCommentRepository(db *sqlx.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) UpsertMany(ctx context.Context, comments []model.Comment) (CommentUpsertResult, error) {
	if len(comments) == 0 {
		return CommentUpsertResult{}, nil
	}

	const query = `
INSERT INTO comments (
    source_id,
    subreddit_id,
    post_id,
    external_id,
    parent_external_id,
    parent_comment_id,
    author_name,
    body,
    score,
    depth,
    published_at
)
VALUES (
    :source_id,
    :subreddit_id,
    :post_id,
    :external_id,
    :parent_external_id,
    :parent_comment_id,
    :author_name,
    :body,
    :score,
    :depth,
    :published_at
)
ON CONFLICT (source_id, external_id)
DO NOTHING;
`

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return CommentUpsertResult{}, fmt.Errorf("begin comment upsert transaction: %w", err)
	}
	defer tx.Rollback()

	inserted := 0
	for _, comment := range comments {
		result, err := tx.NamedExecContext(ctx, query, comment)
		if err != nil {
			return CommentUpsertResult{}, fmt.Errorf("upsert comment %q: %w", comment.ExternalID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return CommentUpsertResult{}, fmt.Errorf("read comment upsert rows affected: %w", err)
		}
		inserted += int(rowsAffected)
	}

	if err := tx.Commit(); err != nil {
		return CommentUpsertResult{}, fmt.Errorf("commit comment upsert transaction: %w", err)
	}

	return CommentUpsertResult{
		Inserted: inserted,
		Skipped:  len(comments) - inserted,
	}, nil
}

func (r *commentRepository) FindByPostID(ctx context.Context, filter CommentFilter) ([]model.Comment, error) {
	query := `
SELECT
    id,
    source_id,
    subreddit_id,
    post_id,
    external_id,
    parent_external_id,
    parent_comment_id,
    author_name,
    body,
    score,
    depth,
    published_at,
    ingested_at,
    updated_at
FROM comments
WHERE post_id = :post_id
`

	args := map[string]any{
		"post_id": filter.PostID,
		"limit":   filter.Limit,
		"offset":  filter.Offset,
	}
	if filter.Depth != nil {
		query += "  AND depth = :depth\n"
		args["depth"] = *filter.Depth
	}
	query += `
ORDER BY published_at ASC
LIMIT :limit
OFFSET :offset;
`

	rows, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("find comments by post id: %w", err)
	}
	defer rows.Close()

	comments := make([]model.Comment, 0)
	for rows.Next() {
		var comment model.Comment
		if err := rows.StructScan(&comment); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}

	return comments, nil
}
