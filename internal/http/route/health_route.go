package route

import (
	"github.com/go-chi/chi/v5"

	"github.com/medivh13/synaptica-api/internal/handler"
)

func RegisterHealthRoutes(r chi.Router, healthHandler *handler.HealthHandler) {
	r.Get("/health", healthHandler.Check)
}
