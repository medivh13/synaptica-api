package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/medivh13/synaptica-api/internal/model"
	"github.com/medivh13/synaptica-api/internal/repository"
)

type AnalyticsService interface {
	GetEmotionTimeline(ctx context.Context, input AnalyticsInput) ([]model.EmotionTimelineBucket, error)
	GetEmotionSummary(ctx context.Context, input AnalyticsInput) (*model.EmotionSummary, error)
	GetTrendingKeywords(ctx context.Context, input AnalyticsInput) ([]model.TrendingKeyword, error)
	GetHotTopics(ctx context.Context, input AnalyticsInput) ([]model.HotTopic, error)
}

type AnalyticsInput struct {
	EntityType string
	Source     string
	PostID     string
	KeywordID  string
	Interval   string
	Limit      int
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository) AnalyticsService {
	return &analyticsService{analyticsRepo: analyticsRepo}
}

func (s *analyticsService) GetEmotionTimeline(ctx context.Context, input AnalyticsInput) ([]model.EmotionTimelineBucket, error) {
	buckets, err := s.analyticsRepo.GetEmotionTimeline(ctx, repository.AnalyticsFilter{
		EntityType: normalizeAnalyticsEntityType(input.EntityType),
		Source:     strings.TrimSpace(strings.ToLower(input.Source)),
		PostID:     strings.TrimSpace(input.PostID),
		KeywordID:  strings.TrimSpace(input.KeywordID),
		Interval:   normalizeAnalyticsInterval(input.Interval),
		Limit:      input.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get emotion timeline: %w", err)
	}
	return buckets, nil
}

func (s *analyticsService) GetEmotionSummary(ctx context.Context, input AnalyticsInput) (*model.EmotionSummary, error) {
	summary, err := s.analyticsRepo.GetEmotionSummary(ctx, repository.AnalyticsFilter{
		EntityType: normalizeAnalyticsEntityType(input.EntityType),
		Source:     strings.TrimSpace(strings.ToLower(input.Source)),
		PostID:     strings.TrimSpace(input.PostID),
		KeywordID:  strings.TrimSpace(input.KeywordID),
		Interval:   normalizeAnalyticsInterval(input.Interval),
		Limit:      input.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get emotion summary: %w", err)
	}
	return summary, nil
}

func (s *analyticsService) GetTrendingKeywords(ctx context.Context, input AnalyticsInput) ([]model.TrendingKeyword, error) {
	keywords, err := s.analyticsRepo.GetTrendingKeywords(ctx, repository.AnalyticsFilter{
		EntityType: normalizeAnalyticsEntityType(input.EntityType),
		Source:     strings.TrimSpace(strings.ToLower(input.Source)),
		Limit:      input.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get trending keywords: %w", err)
	}
	return keywords, nil
}

func (s *analyticsService) GetHotTopics(ctx context.Context, input AnalyticsInput) ([]model.HotTopic, error) {
	topics, err := s.analyticsRepo.GetHotTopics(ctx, repository.AnalyticsFilter{
		Source: strings.TrimSpace(strings.ToLower(input.Source)),
		Limit:  input.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get hot topics: %w", err)
	}
	return topics, nil
}

func normalizeAnalyticsEntityType(entityType string) string {
	entityType = strings.TrimSpace(strings.ToLower(entityType))
	if entityType == "" {
		return "comment"
	}
	return entityType
}

func normalizeAnalyticsInterval(interval string) string {
	interval = strings.TrimSpace(strings.ToLower(interval))
	if interval == "" {
		return "hour"
	}
	return interval
}
