package route

import (
	"github.com/go-chi/chi/v5"

	"github.com/medivh13/synaptica-api/internal/handler"
)

func RegisterSubredditRoutes(r chi.Router, subredditHandler *handler.SubredditHandler) {
	r.Get("/subreddits", subredditHandler.FindAll)
}
