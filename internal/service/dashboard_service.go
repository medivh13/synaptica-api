package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/medivh13/synaptica-api/internal/model"
)

type DashboardService interface {
	GetOverview(ctx context.Context, input DashboardOverviewInput) (*model.DashboardOverview, error)
}

type DashboardOverviewInput struct {
	Source string
	Limit  int
}

type dashboardService struct {
	analyticsService       AnalyticsService
	postAnalysisService    PostAnalysisService
	commentAnalysisService CommentAnalysisService
}

func NewDashboardService(
	analyticsService AnalyticsService,
	postAnalysisService PostAnalysisService,
	commentAnalysisService CommentAnalysisService,
) DashboardService {
	return &dashboardService{
		analyticsService:       analyticsService,
		postAnalysisService:    postAnalysisService,
		commentAnalysisService: commentAnalysisService,
	}
}

func (s *dashboardService) GetOverview(ctx context.Context, input DashboardOverviewInput) (*model.DashboardOverview, error) {
	source := strings.TrimSpace(strings.ToLower(input.Source))

	summary, err := s.analyticsService.GetEmotionSummary(ctx, AnalyticsInput{
		EntityType: "all",
		Source:     source,
	})
	if err != nil {
		return nil, fmt.Errorf("get dashboard summary: %w", err)
	}

	trendingKeywords, err := s.analyticsService.GetTrendingKeywords(ctx, AnalyticsInput{
		EntityType: "all",
		Source:     source,
		Limit:      input.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get dashboard trending keywords: %w", err)
	}

	hotTopics, err := s.analyticsService.GetHotTopics(ctx, AnalyticsInput{
		Source: source,
		Limit:  input.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get dashboard hot topics: %w", err)
	}

	recentPosts, err := s.postAnalysisService.FindAnalyzedPosts(ctx, FindAnalyzedPostsInput{
		Source: source,
		Limit:  input.Limit,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("get dashboard recent posts: %w", err)
	}

	recentComments, err := s.commentAnalysisService.FindAnalyzedComments(ctx, FindAnalyzedCommentsInput{
		Source: source,
		Limit:  input.Limit,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("get dashboard recent comments: %w", err)
	}

	return &model.DashboardOverview{
		Summary:          summary,
		TrendingKeywords: trendingKeywords,
		HotTopics:        hotTopics,
		RecentPosts:      recentPosts,
		RecentComments:   recentComments,
	}, nil
}
