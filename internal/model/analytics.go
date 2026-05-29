package model

import "time"

type EmotionTimelineBucket struct {
	BucketStart           time.Time `json:"bucket_start" db:"bucket_start"`
	BucketEnd             time.Time `json:"bucket_end" db:"bucket_end"`
	TotalItems            int       `json:"total_items" db:"total_items"`
	AvgSentimentScore     float64   `json:"avg_sentiment_score" db:"avg_sentiment_score"`
	AvgToxicityScore      float64   `json:"avg_toxicity_score" db:"avg_toxicity_score"`
	AvgCognitiveLoadScore float64   `json:"avg_cognitive_load_score" db:"avg_cognitive_load_score"`
	AngerCount            int       `json:"anger_count" db:"anger_count"`
	FearCount             int       `json:"fear_count" db:"fear_count"`
	JoyCount              int       `json:"joy_count" db:"joy_count"`
	SadnessCount          int       `json:"sadness_count" db:"sadness_count"`
	DisgustCount          int       `json:"disgust_count" db:"disgust_count"`
	SurpriseCount         int       `json:"surprise_count" db:"surprise_count"`
	NeutralCount          int       `json:"neutral_count" db:"neutral_count"`
}

type EmotionSummary struct {
	TotalItems            int            `json:"total_items" db:"total_items"`
	AvgSentimentScore     float64        `json:"avg_sentiment_score" db:"avg_sentiment_score"`
	AvgToxicityScore      float64        `json:"avg_toxicity_score" db:"avg_toxicity_score"`
	AvgCognitiveLoadScore float64        `json:"avg_cognitive_load_score" db:"avg_cognitive_load_score"`
	DominantEmotion       string         `json:"dominant_emotion"`
	EmotionDistribution   map[string]int `json:"emotion_distribution"`
}

type TrendingKeyword struct {
	KeywordID             string  `json:"keyword_id" db:"keyword_id"`
	Keyword               string  `json:"keyword" db:"keyword"`
	MentionCount          int     `json:"mention_count" db:"mention_count"`
	AvgSentimentScore     float64 `json:"avg_sentiment_score" db:"avg_sentiment_score"`
	AvgToxicityScore      float64 `json:"avg_toxicity_score" db:"avg_toxicity_score"`
	AvgCognitiveLoadScore float64 `json:"avg_cognitive_load_score" db:"avg_cognitive_load_score"`
	DominantEmotion       string  `json:"dominant_emotion" db:"dominant_emotion"`
	TrendScore            float64 `json:"trend_score" db:"trend_score"`
}

type HotTopic struct {
	PostID                  string    `json:"post_id" db:"post_id"`
	Title                   string    `json:"title" db:"title"`
	SourceType              string    `json:"source_type" db:"source_type"`
	Score                   int       `json:"score" db:"score"`
	CommentCount            int       `json:"comment_count" db:"comment_count"`
	PublishedAt             time.Time `json:"published_at" db:"published_at"`
	Permalink               *string   `json:"permalink" db:"permalink,omitempty"`
	TotalAnalyzedComments   int       `json:"total_analyzed_comments" db:"total_analyzed_comments"`
	DominantEmotion         string    `json:"dominant_emotion" db:"dominant_emotion"`
	AvgSentimentScore       float64   `json:"avg_sentiment_score" db:"avg_sentiment_score"`
	AvgToxicityScore        float64   `json:"avg_toxicity_score" db:"avg_toxicity_score"`
	AvgCognitiveLoadScore   float64   `json:"avg_cognitive_load_score" db:"avg_cognitive_load_score"`
	EmotionalIntensityScore float64   `json:"emotional_intensity_score" db:"emotional_intensity_score"`
}
