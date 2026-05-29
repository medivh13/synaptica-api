package route

import (
	"github.com/go-chi/chi/v5"

	"github.com/medivh13/synaptica-api/internal/handler"
)

func RegisterCommentAnalysisRoutes(r chi.Router, commentAnalysisHandler *handler.CommentAnalysisHandler) {
	r.Post("/analysis/comments/run", commentAnalysisHandler.RunCommentAnalysis)
	r.Get("/analysis/comments/results", commentAnalysisHandler.FindCommentAnalysisResults)
}
