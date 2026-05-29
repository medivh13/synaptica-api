package model

type PostInsight struct {
	Post            *Post                   `json:"post"`
	CommentSummary  *EmotionSummary         `json:"comment_summary"`
	CommentTimeline []EmotionTimelineBucket `json:"comment_timeline"`
	TopComments     []AnalyzedComment       `json:"top_comments"`
}
