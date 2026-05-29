package route

import (
	"github.com/go-chi/chi/v5"

	"synaptica-api/internal/handler"
)

func RegisterPostRoutes(r chi.Router, postHandler *handler.PostHandler) {
	r.Get("/posts", postHandler.FindRecent)
}
