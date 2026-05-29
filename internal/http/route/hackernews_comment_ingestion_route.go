package route

import (
	"github.com/go-chi/chi/v5"

	"github.com/medivh13/synaptica-api/internal/handler"
)

func RegisterHackerNewsCommentIngestionRoutes(r chi.Router, hackerNewsCommentIngestionHandler *handler.HackerNewsCommentIngestionHandler) {
	r.Get("/ingestion/hackernews/debug-response", hackerNewsCommentIngestionHandler.DebugResponse)
	r.Post("/ingestion/hackernews/posts/{post_id}/comments", hackerNewsCommentIngestionHandler.IngestPostComments)
}
