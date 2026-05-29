package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	stdhttp "net/http"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"

	"github.com/medivh13/synaptica-api/internal/analyzer"
	"github.com/medivh13/synaptica-api/internal/client/hackernews"
	"github.com/medivh13/synaptica-api/internal/client/reddit"
	"github.com/medivh13/synaptica-api/internal/config"
	"github.com/medivh13/synaptica-api/internal/database"
	"github.com/medivh13/synaptica-api/internal/handler"
	httpserver "github.com/medivh13/synaptica-api/internal/http"
	"github.com/medivh13/synaptica-api/internal/repository"
	"github.com/medivh13/synaptica-api/internal/service"
)

type App struct {
	cfg    config.Config
	db     *sqlx.DB
	server *stdhttp.Server
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	db, err := database.NewPostgres(ctx, database.PostgresConfig{
		DatabaseURL:     cfg.DatabaseURL,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
		ConnMaxIdleTime: cfg.DBConnMaxIdleTime,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize postgres: %w", err)
	}

	subredditRepo := repository.NewSubredditRepository(db)
	dataSourceRepo := repository.NewDataSourceRepository(db)
	postRepo := repository.NewPostRepository(db)
	keywordRepo := repository.NewKeywordRepository(db)
	analysisRepo := repository.NewAnalysisRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	redditClient := reddit.NewClient(cfg.RedditBaseURL, cfg.RedditUserAgent, cfg.RedditRequestTimeout)
	hnClient := hackernews.NewClient(cfg.HNBaseURL, cfg.HNRequestTimeout)
	textAnalyzer := analyzer.NewRuleBasedTextAnalyzer()
	redditIngestionService := service.NewRedditIngestionService(
		redditClient,
		dataSourceRepo,
		subredditRepo,
		postRepo,
		cfg.RedditDefaultLimit,
		cfg.RedditMaxLimit,
	)
	hackerNewsIngestionService := service.NewHackerNewsIngestionService(
		hnClient,
		dataSourceRepo,
		subredditRepo,
		postRepo,
		cfg.HNDefaultLimit,
		cfg.HNMaxLimit,
	)
	hackerNewsCommentIngestionService := service.NewHackerNewsCommentIngestionService(
		hnClient,
		dataSourceRepo,
		postRepo,
		commentRepo,
	)
	keywordService := service.NewKeywordService(keywordRepo)
	postAnalysisService := service.NewPostAnalysisService(analysisRepo, textAnalyzer)
	commentAnalysisService := service.NewCommentAnalysisService(analysisRepo, textAnalyzer)
	analyticsService := service.NewAnalyticsService(analyticsRepo)
	postInsightService := service.NewPostInsightService(postRepo, analyticsService, commentAnalysisService)
	dashboardService := service.NewDashboardService(analyticsService, postAnalysisService, commentAnalysisService)

	responder := httpserver.Responder{}
	healthHandler := handler.NewHealthHandler(cfg.AppName).WithResponder(responder)
	subredditHandler := handler.NewSubredditHandler(subredditRepo).WithResponder(responder)
	postHandler := handler.NewPostHandler(postRepo).WithResponder(responder)
	commentHandler := handler.NewCommentHandler(commentRepo).WithResponder(responder)
	keywordHandler := handler.NewKeywordHandler(keywordService).WithResponder(responder)
	analysisHandler := handler.NewAnalysisHandler(postAnalysisService).WithResponder(responder)
	commentAnalysisHandler := handler.NewCommentAnalysisHandler(commentAnalysisService).WithResponder(responder)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService).WithResponder(responder)
	postInsightHandler := handler.NewPostInsightHandler(postInsightService).WithResponder(responder)
	dashboardHandler := handler.NewDashboardHandler(dashboardService).WithResponder(responder)
	redditIngestionHandler := handler.NewRedditIngestionHandler(redditIngestionService).WithResponder(responder)
	hackerNewsIngestionHandler := handler.NewHackerNewsIngestionHandler(hackerNewsIngestionService).WithResponder(responder)
	hackerNewsCommentIngestionHandler := handler.NewHackerNewsCommentIngestionHandler(hackerNewsCommentIngestionService).WithResponder(responder)
	router := httpserver.NewRouter(httpserver.RouterDependencies{
		HealthHandler:                     healthHandler,
		SubredditHandler:                  subredditHandler,
		PostHandler:                       postHandler,
		CommentHandler:                    commentHandler,
		KeywordHandler:                    keywordHandler,
		AnalysisHandler:                   analysisHandler,
		CommentAnalysisHandler:            commentAnalysisHandler,
		AnalyticsHandler:                  analyticsHandler,
		PostInsightHandler:                postInsightHandler,
		DashboardHandler:                  dashboardHandler,
		RedditIngestionHandler:            redditIngestionHandler,
		HackerNewsIngestionHandler:        hackerNewsIngestionHandler,
		HackerNewsCommentIngestionHandler: hackerNewsCommentIngestionHandler,
	})

	server := &stdhttp.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
		IdleTimeout:  cfg.HTTPIdleTimeout,
	}

	return &App{
		cfg:    cfg,
		db:     db,
		server: server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("starting %s in %s on %s", a.cfg.AppName, a.cfg.AppEnv, a.server.Addr)
		err := a.server.ListenAndServe()
		if err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTPShutdownTimeout)
			defer cancel()
			if shutdownErr := a.Shutdown(shutdownCtx); shutdownErr != nil {
				return fmt.Errorf("server failed: %w; shutdown failed: %v", err, shutdownErr)
			}
			return fmt.Errorf("server failed: %w", err)
		}
		return nil
	case <-ctx.Done():
		log.Printf("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTPShutdownTimeout)
		defer cancel()
		if err := a.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	log.Printf("shutting down http server")
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	log.Printf("closing database connection")
	if err := a.db.Close(); err != nil {
		return fmt.Errorf("close database connection: %w", err)
	}

	log.Printf("shutdown complete")
	return nil
}
