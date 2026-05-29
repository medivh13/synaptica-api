package route

import (
	"github.com/go-chi/chi/v5"

	"synaptica-api/internal/handler"
)

func RegisterDashboardRoutes(r chi.Router, dashboardHandler *handler.DashboardHandler) {
	r.Get("/dashboard/overview", dashboardHandler.GetOverview)
}
