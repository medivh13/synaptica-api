package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"synaptica-api/internal/model"
	"synaptica-api/internal/repository"
)

type PostInsightService interface {
	GetPostInsight(ctx context.Context, postID string) (*model.PostInsight, error)
}

type postInsightService struct {
	postRepo               repository.PostRepository
	analyticsService       AnalyticsService
	commentAnalysisService CommentAnalysisService
}

func NewPostInsightService(
	postRepo repository.PostRepository,
	analyticsService AnalyticsService,
	commentAnalysisService CommentAnalysisService,
) PostInsightService {
	return &postInsightService{
		postRepo:               postRepo,
		analyticsService:       analyticsService,
		commentAnalysisService: commentAnalysisService,
	}
}

func (s *postInsightService) GetPostInsight(ctx context.Context, postID string) (*model.PostInsight, error) {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return nil, fmt.Errorf("%w: post id is required", ErrInvalidInput)
	}

	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: post %q", ErrNotFound, postID)
		}
		return nil, fmt.Errorf("find post: %w", err)
	}

	summary, err := s.analyticsService.GetEmotionSummary(ctx, AnalyticsInput{
		EntityType: "comment",
		PostID:     postID,
	})
	if err != nil {
		return nil, fmt.Errorf("get comment analytics summary: %w", err)
	}

	timeline, err := s.analyticsService.GetEmotionTimeline(ctx, AnalyticsInput{
		EntityType: "comment",
		PostID:     postID,
		Interval:   "hour",
		Limit:      24,
	})
	if err != nil {
		return nil, fmt.Errorf("get comment emotion timeline: %w", err)
	}

	topComments, err := s.commentAnalysisService.FindAnalyzedComments(ctx, FindAnalyzedCommentsInput{
		PostID: postID,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("find analyzed comments: %w", err)
	}

	return &model.PostInsight{
		Post:            post,
		CommentSummary:  summary,
		CommentTimeline: timeline,
		TopComments:     topComments,
	}, nil
}
