package route

import (
	"github.com/go-chi/chi/v5"

	"synaptica-api/internal/handler"
)

func RegisterHackerNewsIngestionRoutes(r chi.Router, hackerNewsIngestionHandler *handler.HackerNewsIngestionHandler) {
	r.Post("/ingestion/hackernews/stories", hackerNewsIngestionHandler.IngestStories)
}
