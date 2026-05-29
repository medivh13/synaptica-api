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

type KeywordService interface {
	Create(ctx context.Context, input KeywordCreateInput) (*model.Keyword, error)
	FindAllActive(ctx context.Context) ([]model.Keyword, error)
	MatchPosts(ctx context.Context, keywordID string) (*KeywordPostMatchResult, error)
	FindPosts(ctx context.Context, keywordID string, limit int, offset int) ([]model.Post, error)
}

type KeywordCreateInput struct {
	Keyword     string
	Description string
}

type KeywordPostMatchResult struct {
	KeywordID string `json:"keyword_id"`
	Keyword   string `json:"keyword"`
	Matched   int    `json:"matched"`
	Inserted  int    `json:"inserted"`
	Skipped   int    `json:"skipped"`
}

type keywordService struct {
	keywordRepo repository.KeywordRepository
}

func NewKeywordService(keywordRepo repository.KeywordRepository) KeywordService {
	return &keywordService{keywordRepo: keywordRepo}
}

func (s *keywordService) Create(ctx context.Context, input KeywordCreateInput) (*model.Keyword, error) {
	keyword := strings.TrimSpace(input.Keyword)
	if keyword == "" {
		return nil, fmt.Errorf("%w: keyword is required", ErrInvalidInput)
	}

	description := stringPtrOrNil(input.Description)
	createdKeyword, err := s.keywordRepo.Create(ctx, repository.KeywordCreateInput{
		Keyword:     keyword,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("create keyword: %w", err)
	}

	return createdKeyword, nil
}

func (s *keywordService) FindAllActive(ctx context.Context) ([]model.Keyword, error) {
	keywords, err := s.keywordRepo.FindAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("find active keywords: %w", err)
	}
	return keywords, nil
}

func (s *keywordService) MatchPosts(ctx context.Context, keywordID string) (*KeywordPostMatchResult, error) {
	keywordID = strings.TrimSpace(keywordID)
	if keywordID == "" {
		return nil, fmt.Errorf("%w: keyword id is required", ErrInvalidInput)
	}

	keyword, err := s.keywordRepo.FindByID(ctx, keywordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: keyword %q", ErrNotFound, keywordID)
		}
		return nil, fmt.Errorf("find keyword: %w", err)
	}

	matchResult, err := s.keywordRepo.MatchPosts(ctx, *keyword)
	if err != nil {
		return nil, fmt.Errorf("match keyword posts: %w", err)
	}

	return &KeywordPostMatchResult{
		KeywordID: keyword.ID,
		Keyword:   keyword.Keyword,
		Matched:   matchResult.Matched,
		Inserted:  matchResult.Inserted,
		Skipped:   matchResult.Skipped,
	}, nil
}

func (s *keywordService) FindPosts(ctx context.Context, keywordID string, limit int, offset int) ([]model.Post, error) {
	keywordID = strings.TrimSpace(keywordID)
	if keywordID == "" {
		return nil, fmt.Errorf("%w: keyword id is required", ErrInvalidInput)
	}

	if _, err := s.keywordRepo.FindByID(ctx, keywordID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: keyword %q", ErrNotFound, keywordID)
		}
		return nil, fmt.Errorf("find keyword: %w", err)
	}

	posts, err := s.keywordRepo.FindPosts(ctx, keywordID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find keyword posts: %w", err)
	}

	return posts, nil
}
