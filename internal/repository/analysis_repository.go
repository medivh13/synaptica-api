package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/medivh13/synaptica-api/internal/model"
)

type AnalysisRepository interface {
	FindUnanalyzedPosts(ctx context.Context, filter AnalysisPostFilter) ([]model.PostAnalysisTarget, error)
	FindUnanalyzedComments(ctx context.Context, filter AnalysisCommentFilter) ([]model.CommentAnalysisTarget, error)
	UpsertAnalysisResults(ctx context.Context, results []model.TextAnalysisResult) (int, error)
	UpsertPostAnalysisResults(ctx context.Context, results []model.TextAnalysisResult) (int, error)
	FindAnalyzedPosts(ctx context.Context, filter AnalysisResultFilter) ([]model.AnalyzedPost, error)
	FindAnalyzedComments(ctx context.Context, filter AnalysisCommentResultFilter) ([]model.AnalyzedComment, error)
}

type AnalysisPostFilter struct {
	Source    string
	KeywordID string
	Limit     int
}

type AnalysisResultFilter struct {
	Source    string
	KeywordID string
	Limit     int
	Offset    int
}

type AnalysisCommentFilter struct {
	PostID    string
	Source    string
	KeywordID string
	Limit     int
}

type AnalysisCommentResultFilter struct {
	PostID    string
	Source    string
	KeywordID string
	Limit     int
	Offset    int
}

type analysisRepository struct {
	db *sqlx.DB
}

func NewAnalysisRepository(db *sqlx.DB) AnalysisRepository {
	return &analysisRepository{db: db}
}

func (r *analysisRepository) FindUnanalyzedPosts(ctx context.Context, filter AnalysisPostFilter) ([]model.PostAnalysisTarget, error) {
	query := `
SELECT posts.id, posts.title, posts.body
FROM posts
JOIN data_sources ds ON ds.id = posts.source_id
LEFT JOIN text_analysis_results tar
  ON tar.entity_type = 'post'
 AND tar.entity_id = posts.id
`
	args := map[string]any{
		"source":     nullableString(filter.Source),
		"keyword_id": nullableString(filter.KeywordID),
		"limit":      filter.Limit,
	}
	if strings.TrimSpace(filter.KeywordID) != "" {
		query += `
JOIN keyword_matches km
  ON km.entity_type = 'post'
 AND km.entity_id = posts.id
 AND km.keyword_id = CAST(:keyword_id AS uuid)
`
	}
	query += `
WHERE tar.id IS NULL
  AND (CAST(:source AS text) IS NULL OR ds.type = CAST(:source AS text))
ORDER BY posts.published_at DESC
LIMIT :limit;
`

	rows, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("find unanalyzed posts: %w", err)
	}
	defer rows.Close()

	posts := make([]model.PostAnalysisTarget, 0)
	for rows.Next() {
		var post model.PostAnalysisTarget
		if err := rows.StructScan(&post); err != nil {
			return nil, fmt.Errorf("scan unanalyzed post: %w", err)
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate unanalyzed posts: %w", err)
	}

	return posts, nil
}

func (r *analysisRepository) UpsertPostAnalysisResults(ctx context.Context, results []model.TextAnalysisResult) (int, error) {
	return r.UpsertAnalysisResults(ctx, results)
}

func (r *analysisRepository) UpsertAnalysisResults(ctx context.Context, results []model.TextAnalysisResult) (int, error) {
	if len(results) == 0 {
		return 0, nil
	}

	const query = `
INSERT INTO text_analysis_results (
    entity_type,
    entity_id,
    sentiment_label,
    sentiment_score,
    emotion_label,
    emotion_score,
    toxicity_score,
    cognitive_load_score,
    language
)
VALUES (
    :entity_type,
    :entity_id,
    :sentiment_label,
    :sentiment_score,
    :emotion_label,
    :emotion_score,
    :toxicity_score,
    :cognitive_load_score,
    :language
)
ON CONFLICT (entity_type, entity_id)
DO UPDATE SET
    sentiment_label = EXCLUDED.sentiment_label,
    sentiment_score = EXCLUDED.sentiment_score,
    emotion_label = EXCLUDED.emotion_label,
    emotion_score = EXCLUDED.emotion_score,
    toxicity_score = EXCLUDED.toxicity_score,
    cognitive_load_score = EXCLUDED.cognitive_load_score,
    language = EXCLUDED.language,
    analyzed_at = NOW();
`

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin analysis upsert transaction: %w", err)
	}
	defer tx.Rollback()

	affected := 0
	for _, result := range results {
		sqlResult, err := tx.NamedExecContext(ctx, query, result)
		if err != nil {
			return 0, fmt.Errorf("upsert analysis result for %s %s: %w", result.EntityType, result.EntityID, err)
		}
		rowsAffected, err := sqlResult.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("read analysis rows affected: %w", err)
		}
		affected += int(rowsAffected)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit analysis upsert transaction: %w", err)
	}

	return affected, nil
}

func (r *analysisRepository) FindUnanalyzedComments(ctx context.Context, filter AnalysisCommentFilter) ([]model.CommentAnalysisTarget, error) {
	query := `
SELECT comments.id, comments.body
FROM comments
JOIN posts ON posts.id = comments.post_id
JOIN data_sources ds ON ds.id = comments.source_id
LEFT JOIN text_analysis_results tar
  ON tar.entity_type = 'comment'
 AND tar.entity_id = comments.id
`
	args := map[string]any{
		"post_id":    nullableString(filter.PostID),
		"source":     nullableString(filter.Source),
		"keyword_id": nullableString(filter.KeywordID),
		"limit":      filter.Limit,
	}
	if strings.TrimSpace(filter.KeywordID) != "" {
		query += `
JOIN keyword_matches km
  ON km.entity_type = 'post'
 AND km.entity_id = posts.id
 AND km.keyword_id = CAST(:keyword_id AS uuid)
`
	}
	query += `
WHERE tar.id IS NULL
  AND (CAST(:post_id AS uuid) IS NULL OR comments.post_id = CAST(:post_id AS uuid))
  AND (CAST(:source AS text) IS NULL OR ds.type = CAST(:source AS text))
ORDER BY comments.published_at ASC
LIMIT :limit;
`

	rows, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("find unanalyzed comments: %w", err)
	}
	defer rows.Close()

	comments := make([]model.CommentAnalysisTarget, 0)
	for rows.Next() {
		var comment model.CommentAnalysisTarget
		if err := rows.StructScan(&comment); err != nil {
			return nil, fmt.Errorf("scan unanalyzed comment: %w", err)
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate unanalyzed comments: %w", err)
	}

	return comments, nil
}

func (r *analysisRepository) FindAnalyzedPosts(ctx context.Context, filter AnalysisResultFilter) ([]model.AnalyzedPost, error) {
	query := `
SELECT
    posts.id,
    posts.title,
    ds.type AS source_type,
    posts.score,
    posts.comment_count,
    posts.published_at,
    posts.permalink,
    tar.sentiment_label,
    tar.sentiment_score,
    tar.emotion_label,
    tar.emotion_score,
    tar.toxicity_score,
    tar.cognitive_load_score,
    tar.analyzed_at
FROM posts
JOIN data_sources ds ON ds.id = posts.source_id
JOIN text_analysis_results tar
  ON tar.entity_type = 'post'
 AND tar.entity_id = posts.id
`
	args := map[string]any{
		"source":     nullableString(filter.Source),
		"keyword_id": nullableString(filter.KeywordID),
		"limit":      filter.Limit,
		"offset":     filter.Offset,
	}
	if strings.TrimSpace(filter.KeywordID) != "" {
		query += `
JOIN keyword_matches km
  ON km.entity_type = 'post'
 AND km.entity_id = posts.id
 AND km.keyword_id = CAST(:keyword_id AS uuid)
`
	}
	query += `
WHERE (CAST(:source AS text) IS NULL OR ds.type = CAST(:source AS text))
ORDER BY posts.published_at DESC
LIMIT :limit
OFFSET :offset;
`

	rows, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("find analyzed posts: %w", err)
	}
	defer rows.Close()

	posts := make([]model.AnalyzedPost, 0)
	for rows.Next() {
		var post model.AnalyzedPost
		if err := rows.StructScan(&post); err != nil {
			return nil, fmt.Errorf("scan analyzed post: %w", err)
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate analyzed posts: %w", err)
	}

	return posts, nil
}

func (r *analysisRepository) FindAnalyzedComments(ctx context.Context, filter AnalysisCommentResultFilter) ([]model.AnalyzedComment, error) {
	query := `
SELECT
    comments.id,
    comments.post_id,
    posts.title AS post_title,
    ds.type AS source_type,
    comments.author_name,
    comments.body,
    comments.depth,
    comments.published_at,
    tar.sentiment_label,
    tar.sentiment_score,
    tar.emotion_label,
    tar.emotion_score,
    tar.toxicity_score,
    tar.cognitive_load_score,
    tar.analyzed_at
FROM comments
JOIN posts ON posts.id = comments.post_id
JOIN data_sources ds ON ds.id = comments.source_id
JOIN text_analysis_results tar
  ON tar.entity_type = 'comment'
 AND tar.entity_id = comments.id
`
	args := map[string]any{
		"post_id":    nullableString(filter.PostID),
		"source":     nullableString(filter.Source),
		"keyword_id": nullableString(filter.KeywordID),
		"limit":      filter.Limit,
		"offset":     filter.Offset,
	}
	if strings.TrimSpace(filter.KeywordID) != "" {
		query += `
JOIN keyword_matches km
  ON km.entity_type = 'post'
 AND km.entity_id = posts.id
 AND km.keyword_id = CAST(:keyword_id AS uuid)
`
	}
	query += `
WHERE (CAST(:post_id AS uuid) IS NULL OR comments.post_id = CAST(:post_id AS uuid))
  AND (CAST(:source AS text) IS NULL OR ds.type = CAST(:source AS text))
ORDER BY comments.published_at ASC
LIMIT :limit
OFFSET :offset;
`

	rows, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("find analyzed comments: %w", err)
	}
	defer rows.Close()

	comments := make([]model.AnalyzedComment, 0)
	for rows.Next() {
		var comment model.AnalyzedComment
		if err := rows.StructScan(&comment); err != nil {
			return nil, fmt.Errorf("scan analyzed comment: %w", err)
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate analyzed comments: %w", err)
	}

	return comments, nil
}
