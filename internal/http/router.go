package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/medivh13/synaptica-api/internal/handler"
	"github.com/medivh13/synaptica-api/internal/http/route"
)

type RouterDependencies struct {
	HealthHandler                     *handler.HealthHandler
	SubredditHandler                  *handler.SubredditHandler
	PostHandler                       *handler.PostHandler
	CommentHandler                    *handler.CommentHandler
	KeywordHandler                    *handler.KeywordHandler
	AnalysisHandler                   *handler.AnalysisHandler
	CommentAnalysisHandler            *handler.CommentAnalysisHandler
	AnalyticsHandler                  *handler.AnalyticsHandler
	PostInsightHandler                *handler.PostInsightHandler
	DashboardHandler                  *handler.DashboardHandler
	RedditIngestionHandler            *handler.RedditIngestionHandler
	HackerNewsIngestionHandler        *handler.HackerNewsIngestionHandler
	HackerNewsCommentIngestionHandler *handler.HackerNewsCommentIngestionHandler
}

func NewRouter(deps RouterDependencies) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(JSONContentType)

	route.RegisterHealthRoutes(r, deps.HealthHandler)

	r.Route("/api/v1", func(api chi.Router) {
		route.RegisterSubredditRoutes(api, deps.SubredditHandler)
		route.RegisterPostRoutes(api, deps.PostHandler)
		route.RegisterCommentRoutes(api, deps.CommentHandler)
		route.RegisterKeywordRoutes(api, deps.KeywordHandler)
		route.RegisterAnalysisRoutes(api, deps.AnalysisHandler)
		route.RegisterCommentAnalysisRoutes(api, deps.CommentAnalysisHandler)
		route.RegisterAnalyticsRoutes(api, deps.AnalyticsHandler)
		route.RegisterPostInsightRoutes(api, deps.PostInsightHandler)
		route.RegisterDashboardRoutes(api, deps.DashboardHandler)
		route.RegisterIngestionRoutes(api, deps.RedditIngestionHandler)
		route.RegisterHackerNewsIngestionRoutes(api, deps.HackerNewsIngestionHandler)
		route.RegisterHackerNewsCommentIngestionRoutes(api, deps.HackerNewsCommentIngestionHandler)
	})

	return r
}
