package route

import (
	"github.com/go-chi/chi/v5"

	"github.com/medivh13/synaptica-api/internal/handler"
)

func RegisterAnalysisRoutes(r chi.Router, analysisHandler *handler.AnalysisHandler) {
	r.Post("/analysis/posts/run", analysisHandler.RunPostAnalysis)
	r.Get("/analysis/posts/results", analysisHandler.FindPostAnalysisResults)
}
