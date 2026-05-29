package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"synaptica-api/internal/client/reddit"
	"synaptica-api/internal/model"
	"synaptica-api/internal/repository"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
)

type RedditIngestionService interface {
	IngestLatestPosts(ctx context.Context, subredditName string, limit int) (*RedditPostIngestionResult, error)
}

type RedditPostIngestionResult struct {
	Subreddit string `json:"subreddit"`
	Fetched   int    `json:"fetched"`
	Inserted  int    `json:"inserted"`
	Skipped   int    `json:"skipped"`
}

type redditIngestionService struct {
	redditClient   reddit.Client
	dataSourceRepo repository.DataSourceRepository
	subredditRepo  repository.SubredditRepository
	postRepo       repository.PostRepository
	defaultLimit   int
	maxLimit       int
}

func NewRedditIngestionService(
	redditClient reddit.Client,
	dataSourceRepo repository.DataSourceRepository,
	subredditRepo repository.SubredditRepository,
	postRepo repository.PostRepository,
	defaultLimit int,
	maxLimit int,
) RedditIngestionService {
	return &redditIngestionService{
		redditClient:   redditClient,
		dataSourceRepo: dataSourceRepo,
		subredditRepo:  subredditRepo,
		postRepo:       postRepo,
		defaultLimit:   defaultLimit,
		maxLimit:       maxLimit,
	}
}

func (s *redditIngestionService) IngestLatestPosts(ctx context.Context, subredditName string, limit int) (*RedditPostIngestionResult, error) {
	subredditName = sanitizeSubredditName(subredditName)
	if subredditName == "" {
		return nil, fmt.Errorf("%w: subreddit is required", ErrInvalidInput)
	}

	if limit <= 0 {
		limit = s.defaultLimit
	}
	if limit > s.maxLimit {
		limit = s.maxLimit
	}

	dataSource, err := s.dataSourceRepo.FindByType(ctx, "reddit")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: reddit data source", ErrNotFound)
		}
		return nil, fmt.Errorf("find reddit data source: %w", err)
	}

	subreddit, err := s.subredditRepo.FindByName(ctx, subredditName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: subreddit %q", ErrNotFound, subredditName)
		}
		return nil, fmt.Errorf("find subreddit: %w", err)
	}

	redditPosts, err := s.redditClient.FetchLatestPosts(ctx, subredditName, limit)
	if err != nil {
		return nil, fmt.Errorf("fetch reddit posts: %w", err)
	}

	posts := make([]model.Post, 0, len(redditPosts))
	for _, redditPost := range redditPosts {
		posts = append(posts, mapRedditPost(dataSource.ID, subreddit.ID, redditPost))
	}

	upsertResult, err := s.postRepo.UpsertMany(ctx, posts)
	if err != nil {
		return nil, fmt.Errorf("store reddit posts: %w", err)
	}

	return &RedditPostIngestionResult{
		Subreddit: subredditName,
		Fetched:   len(redditPosts),
		Inserted:  upsertResult.Inserted,
		Skipped:   upsertResult.Skipped,
	}, nil
}

func sanitizeSubredditName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "/")
	name = strings.TrimPrefix(strings.ToLower(name), "r/")
	return strings.TrimSpace(name)
}

func mapRedditPost(sourceID string, subredditID string, redditPost reddit.Post) model.Post {
	return model.Post{
		SourceID:     sourceID,
		SubredditID:  subredditID,
		ExternalID:   redditPost.ID,
		Title:        redditPost.Title,
		Body:         stringPtrOrNil(redditPost.SelfText),
		AuthorName:   stringPtrOrNil(redditPost.Author),
		Score:        redditPost.Score,
		UpvoteRatio:  floatPtrOrNil(redditPost.UpvoteRatio),
		CommentCount: redditPost.NumComments,
		Permalink:    stringPtrOrNil(redditPost.Permalink),
		URL:          stringPtrOrNil(redditPost.URL),
		PublishedAt:  time.Unix(redditPost.CreatedUTC, 0).UTC(),
	}
}

func stringPtrOrNil(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func floatPtrOrNil(value float64) *float64 {
	if value == 0 {
		return nil
	}
	return &value
}
