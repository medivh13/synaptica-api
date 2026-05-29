package model

import "time"

type TextAnalysisResult struct {
	ID                 string    `db:"id" json:"id"`
	EntityType         string    `db:"entity_type" json:"entity_type"`
	EntityID           string    `db:"entity_id" json:"entity_id"`
	SentimentLabel     string    `db:"sentiment_label" json:"sentiment_label"`
	SentimentScore     float64   `db:"sentiment_score" json:"sentiment_score"`
	EmotionLabel       string    `db:"emotion_label" json:"emotion_label"`
	EmotionScore       float64   `db:"emotion_score" json:"emotion_score"`
	ToxicityScore      float64   `db:"toxicity_score" json:"toxicity_score"`
	CognitiveLoadScore float64   `db:"cognitive_load_score" json:"cognitive_load_score"`
	Language           string    `db:"language" json:"language"`
	AnalyzedAt         time.Time `db:"analyzed_at" json:"analyzed_at"`
}

type PostAnalysisTarget struct {
	ID    string  `db:"id"`
	Title string  `db:"title"`
	Body  *string `db:"body"`
}

type CommentAnalysisTarget struct {
	ID   string `db:"id"`
	Body string `db:"body"`
}

type AnalyzedPost struct {
	ID                 string    `db:"id" json:"id"`
	Title              string    `db:"title" json:"title"`
	SourceType         string    `db:"source_type" json:"source_type"`
	Score              int       `db:"score" json:"score"`
	CommentCount       int       `db:"comment_count" json:"comment_count"`
	PublishedAt        time.Time `db:"published_at" json:"published_at"`
	Permalink          *string   `db:"permalink" json:"permalink,omitempty"`
	SentimentLabel     string    `db:"sentiment_label" json:"sentiment_label"`
	SentimentScore     float64   `db:"sentiment_score" json:"sentiment_score"`
	EmotionLabel       string    `db:"emotion_label" json:"emotion_label"`
	EmotionScore       float64   `db:"emotion_score" json:"emotion_score"`
	ToxicityScore      float64   `db:"toxicity_score" json:"toxicity_score"`
	CognitiveLoadScore float64   `db:"cognitive_load_score" json:"cognitive_load_score"`
	AnalyzedAt         time.Time `db:"analyzed_at" json:"analyzed_at"`
}

type AnalyzedComment struct {
	ID                 string    `db:"id" json:"id"`
	PostID             string    `db:"post_id" json:"post_id"`
	PostTitle          string    `db:"post_title" json:"post_title"`
	SourceType         string    `db:"source_type" json:"source_type"`
	AuthorName         *string   `db:"author_name" json:"author_name,omitempty"`
	Body               string    `db:"body" json:"body"`
	Depth              int       `db:"depth" json:"depth"`
	PublishedAt        time.Time `db:"published_at" json:"published_at"`
	SentimentLabel     string    `db:"sentiment_label" json:"sentiment_label"`
	SentimentScore     float64   `db:"sentiment_score" json:"sentiment_score"`
	EmotionLabel       string    `db:"emotion_label" json:"emotion_label"`
	EmotionScore       float64   `db:"emotion_score" json:"emotion_score"`
	ToxicityScore      float64   `db:"toxicity_score" json:"toxicity_score"`
	CognitiveLoadScore float64   `db:"cognitive_load_score" json:"cognitive_load_score"`
	AnalyzedAt         time.Time `db:"analyzed_at" json:"analyzed_at"`
}
