package route

import (
	"github.com/go-chi/chi/v5"

	"github.com/medivh13/synaptica-api/internal/handler"
)

func RegisterAnalyticsRoutes(r chi.Router, analyticsHandler *handler.AnalyticsHandler) {
	r.Get("/analytics/emotions/timeline", analyticsHandler.GetEmotionTimeline)
	r.Get("/analytics/emotions/summary", analyticsHandler.GetEmotionSummary)
	r.Get("/analytics/trending-keywords", analyticsHandler.GetTrendingKeywords)
	r.Get("/analytics/hot-topics", analyticsHandler.GetHotTopics)
}
