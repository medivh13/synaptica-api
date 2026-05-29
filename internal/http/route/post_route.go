package route

import (
	"github.com/go-chi/chi/v5"

	"github.com/medivh13/synaptica-api/internal/handler"
)

func RegisterPostRoutes(r chi.Router, postHandler *handler.PostHandler) {
	r.Get("/posts", postHandler.FindRecent)
}
