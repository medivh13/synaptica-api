package route

import (
	"github.com/go-chi/chi/v5"

	"github.com/medivh13/synaptica-api/internal/handler"
)

func RegisterKeywordRoutes(r chi.Router, keywordHandler *handler.KeywordHandler) {
	r.Post("/keywords", keywordHandler.Create)
	r.Get("/keywords", keywordHandler.FindAllActive)
	r.Post("/keywords/{keyword_id}/match-posts", keywordHandler.MatchPosts)
	r.Get("/keywords/{keyword_id}/posts", keywordHandler.FindPosts)
}
