package model

type DashboardOverview struct {
	Summary          *EmotionSummary   `json:"summary"`
	TrendingKeywords []TrendingKeyword `json:"trending_keywords"`
	HotTopics        []HotTopic        `json:"hot_topics"`
	RecentPosts      []AnalyzedPost    `json:"recent_posts"`
	RecentComments   []AnalyzedComment `json:"recent_comments"`
}
