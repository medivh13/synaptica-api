package route

import (
	"github.com/go-chi/chi/v5"

	"synaptica-api/internal/handler"
)

func RegisterIngestionRoutes(r chi.Router, redditIngestionHandler *handler.RedditIngestionHandler) {
	r.Post("/ingestion/reddit/subreddits/{subreddit}/posts", redditIngestionHandler.IngestLatestPosts)
}
