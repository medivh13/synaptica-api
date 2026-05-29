package route

import (
	"github.com/go-chi/chi/v5"

	"synaptica-api/internal/handler"
)

func RegisterPostInsightRoutes(r chi.Router, postInsightHandler *handler.PostInsightHandler) {
	r.Get("/insights/posts/{post_id}", postInsightHandler.GetPostInsight)
}
