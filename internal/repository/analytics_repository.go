package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/medivh13/synaptica-api/internal/model"
)

type AnalyticsRepository interface {
	GetEmotionTimeline(ctx context.Context, filter AnalyticsFilter) ([]model.EmotionTimelineBucket, error)
	GetEmotionSummary(ctx context.Context, filter AnalyticsFilter) (*model.EmotionSummary, error)
	GetTrendingKeywords(ctx context.Context, filter AnalyticsFilter) ([]model.TrendingKeyword, error)
	GetHotTopics(ctx context.Context, filter AnalyticsFilter) ([]model.HotTopic, error)
}

type AnalyticsFilter struct {
	EntityType string
	Source     string
	PostID     string
	KeywordID  string
	Interval   string
	Limit      int
}

type analyticsRepository struct {
	db *sqlx.DB
}

type emotionSummaryRow struct {
	TotalItems            int     `db:"total_items"`
	AvgSentimentScore     float64 `db:"avg_sentiment_score"`
	AvgToxicityScore      float64 `db:"avg_toxicity_score"`
	AvgCognitiveLoadScore float64 `db:"avg_cognitive_load_score"`
	AngerCount            int     `db:"anger_count"`
	FearCount             int     `db:"fear_count"`
	JoyCount              int     `db:"joy_count"`
	SadnessCount          int     `db:"sadness_count"`
	DisgustCount          int     `db:"disgust_count"`
	SurpriseCount         int     `db:"surprise_count"`
	NeutralCount          int     `db:"neutral_count"`
}

func NewAnalyticsRepository(db *sqlx.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) GetEmotionTimeline(ctx context.Context, filter AnalyticsFilter) ([]model.EmotionTimelineBucket, error) {
	baseQuery := r.buildAnalyticsBaseQuery(filter.EntityType)
	query := `
WITH analysis_items AS (
` + baseQuery + `
),
bucketed AS (
    SELECT
        date_trunc(CAST(:interval AS text), published_at) AS bucket_start,
        date_trunc(CAST(:interval AS text), published_at) +
            CASE WHEN CAST(:interval AS text) = 'hour' THEN INTERVAL '1 hour' ELSE INTERVAL '1 day' END AS bucket_end,
        CAST(COUNT(*) AS int) AS total_items,
        CAST(COALESCE(AVG(sentiment_score), 0) AS double precision) AS avg_sentiment_score,
        CAST(COALESCE(AVG(toxicity_score), 0) AS double precision) AS avg_toxicity_score,
        CAST(COALESCE(AVG(cognitive_load_score), 0) AS double precision) AS avg_cognitive_load_score,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'anger') AS int) AS anger_count,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'fear') AS int) AS fear_count,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'joy') AS int) AS joy_count,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'sadness') AS int) AS sadness_count,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'disgust') AS int) AS disgust_count,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'surprise') AS int) AS surprise_count,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'neutral') AS int) AS neutral_count
    FROM analysis_items
    GROUP BY bucket_start, bucket_end
    ORDER BY bucket_start DESC
    LIMIT :limit
)
SELECT *
FROM bucketed
ORDER BY bucket_start ASC;
`

	rows, err := r.db.NamedQueryContext(ctx, query, analyticsArgs(filter))
	if err != nil {
		return nil, fmt.Errorf("get emotion timeline: %w", err)
	}
	defer rows.Close()

	buckets := make([]model.EmotionTimelineBucket, 0)
	for rows.Next() {
		var bucket model.EmotionTimelineBucket
		if err := rows.StructScan(&bucket); err != nil {
			return nil, fmt.Errorf("scan emotion timeline bucket: %w", err)
		}
		buckets = append(buckets, bucket)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate emotion timeline buckets: %w", err)
	}

	return buckets, nil
}

func (r *analyticsRepository) GetEmotionSummary(ctx context.Context, filter AnalyticsFilter) (*model.EmotionSummary, error) {
	baseQuery := r.buildAnalyticsBaseQuery(filter.EntityType)
	query := `
WITH analysis_items AS (
` + baseQuery + `
)
SELECT
    CAST(COUNT(*) AS int) AS total_items,
    CAST(COALESCE(AVG(sentiment_score), 0) AS double precision) AS avg_sentiment_score,
    CAST(COALESCE(AVG(toxicity_score), 0) AS double precision) AS avg_toxicity_score,
    CAST(COALESCE(AVG(cognitive_load_score), 0) AS double precision) AS avg_cognitive_load_score,
    CAST(COUNT(*) FILTER (WHERE emotion_label = 'anger') AS int) AS anger_count,
    CAST(COUNT(*) FILTER (WHERE emotion_label = 'fear') AS int) AS fear_count,
    CAST(COUNT(*) FILTER (WHERE emotion_label = 'joy') AS int) AS joy_count,
    CAST(COUNT(*) FILTER (WHERE emotion_label = 'sadness') AS int) AS sadness_count,
    CAST(COUNT(*) FILTER (WHERE emotion_label = 'disgust') AS int) AS disgust_count,
    CAST(COUNT(*) FILTER (WHERE emotion_label = 'surprise') AS int) AS surprise_count,
    CAST(COUNT(*) FILTER (WHERE emotion_label = 'neutral') AS int) AS neutral_count
FROM analysis_items;
`

	stmt, args, err := sqlx.Named(query, analyticsArgs(filter))
	if err != nil {
		return nil, fmt.Errorf("bind emotion summary query: %w", err)
	}
	stmt = r.db.Rebind(stmt)

	var row emotionSummaryRow
	if err := r.db.GetContext(ctx, &row, stmt, args...); err != nil {
		return nil, fmt.Errorf("get emotion summary: %w", err)
	}

	distribution := map[string]int{
		"anger":    row.AngerCount,
		"fear":     row.FearCount,
		"joy":      row.JoyCount,
		"sadness":  row.SadnessCount,
		"disgust":  row.DisgustCount,
		"surprise": row.SurpriseCount,
		"neutral":  row.NeutralCount,
	}

	return &model.EmotionSummary{
		TotalItems:            row.TotalItems,
		AvgSentimentScore:     row.AvgSentimentScore,
		AvgToxicityScore:      row.AvgToxicityScore,
		AvgCognitiveLoadScore: row.AvgCognitiveLoadScore,
		DominantEmotion:       dominantEmotion(distribution),
		EmotionDistribution:   distribution,
	}, nil
}

func (r *analyticsRepository) GetTrendingKeywords(ctx context.Context, filter AnalyticsFilter) ([]model.TrendingKeyword, error) {
	baseQuery := r.buildTrendingKeywordBaseQuery(filter.EntityType)
	query := `
WITH keyword_activity AS (
` + baseQuery + `
),
keyword_aggregates AS (
    SELECT
        keyword_id,
        keyword,
        CAST(COUNT(*) AS int) AS mention_count,
        CAST(COALESCE(AVG(sentiment_score), 0) AS double precision) AS avg_sentiment_score,
        CAST(COALESCE(AVG(toxicity_score), 0) AS double precision) AS avg_toxicity_score,
        CAST(COALESCE(AVG(cognitive_load_score), 0) AS double precision) AS avg_cognitive_load_score,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'anger') AS int) AS anger_count,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'fear') AS int) AS fear_count,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'joy') AS int) AS joy_count,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'sadness') AS int) AS sadness_count,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'disgust') AS int) AS disgust_count,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'surprise') AS int) AS surprise_count,
        CAST(COUNT(*) FILTER (WHERE emotion_label = 'neutral') AS int) AS neutral_count
    FROM keyword_activity
    GROUP BY keyword_id, keyword
)
SELECT
    keyword_id,
    keyword,
    mention_count,
    avg_sentiment_score,
    avg_toxicity_score,
    avg_cognitive_load_score,
    CASE
        WHEN anger_count >= fear_count AND anger_count >= joy_count AND anger_count >= sadness_count AND anger_count >= disgust_count AND anger_count >= surprise_count AND anger_count >= neutral_count THEN 'anger'
        WHEN fear_count >= joy_count AND fear_count >= sadness_count AND fear_count >= disgust_count AND fear_count >= surprise_count AND fear_count >= neutral_count THEN 'fear'
        WHEN joy_count >= sadness_count AND joy_count >= disgust_count AND joy_count >= surprise_count AND joy_count >= neutral_count THEN 'joy'
        WHEN sadness_count >= disgust_count AND sadness_count >= surprise_count AND sadness_count >= neutral_count THEN 'sadness'
        WHEN disgust_count >= surprise_count AND disgust_count >= neutral_count THEN 'disgust'
        WHEN surprise_count >= neutral_count THEN 'surprise'
        ELSE 'neutral'
    END AS dominant_emotion,
    CAST((mention_count + (avg_toxicity_score * 10) + (avg_cognitive_load_score * 5)) AS double precision) AS trend_score
FROM keyword_aggregates
ORDER BY trend_score DESC, mention_count DESC, keyword ASC
LIMIT :limit;
`

	rows, err := r.db.NamedQueryContext(ctx, query, analyticsArgs(filter))
	if err != nil {
		return nil, fmt.Errorf("get trending keywords: %w", err)
	}
	defer rows.Close()

	keywords := make([]model.TrendingKeyword, 0)
	for rows.Next() {
		var keyword model.TrendingKeyword
		if err := rows.StructScan(&keyword); err != nil {
			return nil, fmt.Errorf("scan trending keyword: %w", err)
		}
		keywords = append(keywords, keyword)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate trending keywords: %w", err)
	}

	return keywords, nil
}

func (r *analyticsRepository) GetHotTopics(ctx context.Context, filter AnalyticsFilter) ([]model.HotTopic, error) {
	const query = `
WITH hot_topic_aggregates AS (
    SELECT
        posts.id AS post_id,
        posts.title,
        ds.type AS source_type,
        posts.score,
        posts.comment_count,
        posts.published_at,
        posts.permalink,
        CAST(COUNT(*) AS int) AS total_analyzed_comments,
        CAST(COALESCE(AVG(tar.sentiment_score), 0) AS double precision) AS avg_sentiment_score,
        CAST(COALESCE(AVG(tar.toxicity_score), 0) AS double precision) AS avg_toxicity_score,
        CAST(COALESCE(AVG(tar.cognitive_load_score), 0) AS double precision) AS avg_cognitive_load_score,
        CAST(COUNT(*) FILTER (WHERE tar.emotion_label = 'anger') AS int) AS anger_count,
        CAST(COUNT(*) FILTER (WHERE tar.emotion_label = 'fear') AS int) AS fear_count,
        CAST(COUNT(*) FILTER (WHERE tar.emotion_label = 'joy') AS int) AS joy_count,
        CAST(COUNT(*) FILTER (WHERE tar.emotion_label = 'sadness') AS int) AS sadness_count,
        CAST(COUNT(*) FILTER (WHERE tar.emotion_label = 'disgust') AS int) AS disgust_count,
        CAST(COUNT(*) FILTER (WHERE tar.emotion_label = 'surprise') AS int) AS surprise_count,
        CAST(COUNT(*) FILTER (WHERE tar.emotion_label = 'neutral') AS int) AS neutral_count
    FROM posts
    JOIN data_sources ds ON ds.id = posts.source_id
    JOIN comments ON comments.post_id = posts.id
    JOIN text_analysis_results tar
      ON tar.entity_type = 'comment'
     AND tar.entity_id = comments.id
    WHERE (CAST(:source AS text) IS NULL OR ds.type = CAST(:source AS text))
    GROUP BY posts.id, posts.title, ds.type, posts.score, posts.comment_count, posts.published_at, posts.permalink
)
SELECT
    post_id,
    title,
    source_type,
    score,
    comment_count,
    published_at,
    permalink,
    total_analyzed_comments,
    CASE
        WHEN anger_count >= fear_count AND anger_count >= joy_count AND anger_count >= sadness_count AND anger_count >= disgust_count AND anger_count >= surprise_count AND anger_count >= neutral_count THEN 'anger'
        WHEN fear_count >= joy_count AND fear_count >= sadness_count AND fear_count >= disgust_count AND fear_count >= surprise_count AND fear_count >= neutral_count THEN 'fear'
        WHEN joy_count >= sadness_count AND joy_count >= disgust_count AND joy_count >= surprise_count AND joy_count >= neutral_count THEN 'joy'
        WHEN sadness_count >= disgust_count AND sadness_count >= surprise_count AND sadness_count >= neutral_count THEN 'sadness'
        WHEN disgust_count >= surprise_count AND disgust_count >= neutral_count THEN 'disgust'
        WHEN surprise_count >= neutral_count THEN 'surprise'
        ELSE 'neutral'
    END AS dominant_emotion,
    avg_sentiment_score,
    avg_toxicity_score,
    avg_cognitive_load_score,
    CAST((
        joy_count + fear_count + anger_count + sadness_count + disgust_count + surprise_count
        + (avg_toxicity_score * 10)
        + (avg_cognitive_load_score * 5)
    ) AS double precision) AS emotional_intensity_score
FROM hot_topic_aggregates
ORDER BY emotional_intensity_score DESC, total_analyzed_comments DESC, published_at DESC
LIMIT :limit;
`

	rows, err := r.db.NamedQueryContext(ctx, query, analyticsArgs(filter))
	if err != nil {
		return nil, fmt.Errorf("get hot topics: %w", err)
	}
	defer rows.Close()

	topics := make([]model.HotTopic, 0)
	for rows.Next() {
		var topic model.HotTopic
		if err := rows.StructScan(&topic); err != nil {
			return nil, fmt.Errorf("scan hot topic: %w", err)
		}
		topics = append(topics, topic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate hot topics: %w", err)
	}

	return topics, nil
}

func (r *analyticsRepository) buildAnalyticsBaseQuery(entityType string) string {
	switch entityType {
	case "post":
		return postAnalyticsSelect()
	case "all":
		return postAnalyticsSelect() + "\nUNION ALL\n" + commentAnalyticsSelect()
	default:
		return commentAnalyticsSelect()
	}
}

func (r *analyticsRepository) buildTrendingKeywordBaseQuery(entityType string) string {
	switch entityType {
	case "post":
		return postKeywordActivitySelect()
	case "comment":
		return commentKeywordActivitySelect()
	default:
		return postKeywordActivitySelect() + "\nUNION ALL\n" + commentKeywordActivitySelect()
	}
}

func postKeywordActivitySelect() string {
	return `
SELECT
    keywords.id AS keyword_id,
    keywords.keyword,
    tar.sentiment_score,
    tar.emotion_label,
    tar.toxicity_score,
    tar.cognitive_load_score
FROM keyword_matches km
JOIN keywords ON keywords.id = km.keyword_id
JOIN posts ON km.entity_type = 'post' AND km.entity_id = posts.id
JOIN data_sources ds ON ds.id = posts.source_id
JOIN text_analysis_results tar ON tar.entity_type = 'post' AND tar.entity_id = posts.id
WHERE (CAST(:source AS text) IS NULL OR ds.type = CAST(:source AS text))
`
}

func commentKeywordActivitySelect() string {
	return `
SELECT
    keywords.id AS keyword_id,
    keywords.keyword,
    tar.sentiment_score,
    tar.emotion_label,
    tar.toxicity_score,
    tar.cognitive_load_score
FROM keyword_matches km
JOIN keywords ON keywords.id = km.keyword_id
JOIN posts ON km.entity_type = 'post' AND km.entity_id = posts.id
JOIN data_sources ds ON ds.id = posts.source_id
JOIN comments ON comments.post_id = posts.id
JOIN text_analysis_results tar ON tar.entity_type = 'comment' AND tar.entity_id = comments.id
WHERE (CAST(:source AS text) IS NULL OR ds.type = CAST(:source AS text))
`
}

func postAnalyticsSelect() string {
	return `
SELECT
    tar.entity_type,
    posts.id AS post_id,
    posts.source_id,
    posts.published_at,
    tar.sentiment_score,
    tar.emotion_label,
    tar.toxicity_score,
    tar.cognitive_load_score
FROM text_analysis_results tar
JOIN posts ON tar.entity_type = 'post' AND tar.entity_id = posts.id
JOIN data_sources ds ON ds.id = posts.source_id
WHERE (CAST(:source AS text) IS NULL OR ds.type = CAST(:source AS text))
  AND (CAST(:post_id AS uuid) IS NULL OR posts.id = CAST(:post_id AS uuid))
  AND (
    CAST(:keyword_id AS uuid) IS NULL
    OR EXISTS (
        SELECT 1
        FROM keyword_matches km
        WHERE km.keyword_id = CAST(:keyword_id AS uuid)
          AND km.entity_type = 'post'
          AND km.entity_id = posts.id
    )
  )
`
}

func commentAnalyticsSelect() string {
	return `
SELECT
    tar.entity_type,
    comments.post_id,
    comments.source_id,
    comments.published_at,
    tar.sentiment_score,
    tar.emotion_label,
    tar.toxicity_score,
    tar.cognitive_load_score
FROM text_analysis_results tar
JOIN comments ON tar.entity_type = 'comment' AND tar.entity_id = comments.id
JOIN posts ON posts.id = comments.post_id
JOIN data_sources ds ON ds.id = comments.source_id
WHERE (CAST(:source AS text) IS NULL OR ds.type = CAST(:source AS text))
  AND (CAST(:post_id AS uuid) IS NULL OR comments.post_id = CAST(:post_id AS uuid))
  AND (
    CAST(:keyword_id AS uuid) IS NULL
    OR EXISTS (
        SELECT 1
        FROM keyword_matches km
        WHERE km.keyword_id = CAST(:keyword_id AS uuid)
          AND km.entity_type = 'post'
          AND km.entity_id = posts.id
    )
  )
`
}

func analyticsArgs(filter AnalyticsFilter) map[string]any {
	return map[string]any{
		"source":     nullableString(filter.Source),
		"post_id":    nullableString(filter.PostID),
		"keyword_id": nullableString(filter.KeywordID),
		"interval":   filter.Interval,
		"limit":      filter.Limit,
	}
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func dominantEmotion(distribution map[string]int) string {
	order := []string{"anger", "fear", "joy", "sadness", "disgust", "surprise", "neutral"}
	dominant := "neutral"
	maxCount := 0
	for _, emotion := range order {
		if distribution[emotion] > maxCount {
			dominant = emotion
			maxCount = distribution[emotion]
		}
	}
	return dominant
}
