package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"synaptica-api/internal/client/hackernews"
	"synaptica-api/internal/model"
	"synaptica-api/internal/repository"
)

type HackerNewsIngestionService interface {
	IngestStories(ctx context.Context, feed string, limit int) (*HackerNewsIngestionResult, error)
}

type HackerNewsIngestionResult struct {
	Source   string `json:"source"`
	Feed     string `json:"feed"`
	Fetched  int    `json:"fetched"`
	Inserted int    `json:"inserted"`
	Skipped  int    `json:"skipped"`
}

type hackerNewsIngestionService struct {
	hnClient       hackernews.Client
	dataSourceRepo repository.DataSourceRepository
	subredditRepo  repository.SubredditRepository
	postRepo       repository.PostRepository
	defaultLimit   int
	maxLimit       int
}

var allowedHNFeeds = map[string]struct{}{
	"topstories":  {},
	"newstories":  {},
	"beststories": {},
	"askstories":  {},
	"showstories": {},
	"jobstories":  {},
}

func NewHackerNewsIngestionService(
	hnClient hackernews.Client,
	dataSourceRepo repository.DataSourceRepository,
	subredditRepo repository.SubredditRepository,
	postRepo repository.PostRepository,
	defaultLimit int,
	maxLimit int,
) HackerNewsIngestionService {
	return &hackerNewsIngestionService{
		hnClient:       hnClient,
		dataSourceRepo: dataSourceRepo,
		subredditRepo:  subredditRepo,
		postRepo:       postRepo,
		defaultLimit:   defaultLimit,
		maxLimit:       maxLimit,
	}
}

func (s *hackerNewsIngestionService) IngestStories(ctx context.Context, feed string, limit int) (*HackerNewsIngestionResult, error) {
	feed = normalizeHNFeed(feed)
	if _, ok := allowedHNFeeds[feed]; !ok {
		return nil, fmt.Errorf("%w: unsupported hacker news feed %q", ErrInvalidInput, feed)
	}

	if limit <= 0 {
		limit = s.defaultLimit
	}
	if limit > s.maxLimit {
		limit = s.maxLimit
	}

	dataSource, err := s.dataSourceRepo.FindByType(ctx, "hackernews")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: hackernews data source", ErrNotFound)
		}
		return nil, fmt.Errorf("find hackernews data source: %w", err)
	}

	subreddit, err := s.subredditRepo.FindOrCreateHN(ctx, dataSource.ID)
	if err != nil {
		return nil, fmt.Errorf("find or create hackernews subreddit: %w", err)
	}

	storyIDs, err := s.hnClient.FetchStoryIDs(ctx, feed)
	if err != nil {
		return nil, fmt.Errorf("fetch hackernews story ids: %w", err)
	}
	if len(storyIDs) > limit {
		storyIDs = storyIDs[:limit]
	}

	posts := make([]model.Post, 0, len(storyIDs))
	for _, id := range storyIDs {
		item, err := s.hnClient.FetchItem(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("fetch hackernews item %d: %w", id, err)
		}
		if shouldSkipHNItem(item) {
			continue
		}
		posts = append(posts, mapHNItem(dataSource.ID, subreddit.ID, item))
	}

	upsertResult, err := s.postRepo.UpsertMany(ctx, posts)
	if err != nil {
		return nil, fmt.Errorf("store hackernews posts: %w", err)
	}

	return &HackerNewsIngestionResult{
		Source:   "hackernews",
		Feed:     feed,
		Fetched:  len(posts),
		Inserted: upsertResult.Inserted,
		Skipped:  upsertResult.Skipped,
	}, nil
}

func normalizeHNFeed(feed string) string {
	feed = strings.TrimSpace(strings.ToLower(feed))
	if feed == "" {
		return "topstories"
	}
	return feed
}

func shouldSkipHNItem(item *hackernews.Item) bool {
	if item == nil || item.Deleted || item.Dead {
		return true
	}
	return item.Type != "story" && item.Type != "job"
}

func mapHNItem(sourceID string, subredditID string, item *hackernews.Item) model.Post {
	permalink := fmt.Sprintf("https://news.ycombinator.com/item?id=%d", item.ID)
	return model.Post{
		SourceID:     sourceID,
		SubredditID:  subredditID,
		ExternalID:   strconv.FormatInt(item.ID, 10),
		Title:        item.Title,
		Body:         stringPtrOrNil(item.Text),
		AuthorName:   stringPtrOrNil(item.By),
		Score:        item.Score,
		UpvoteRatio:  nil,
		CommentCount: item.Descendants,
		Permalink:    &permalink,
		URL:          stringPtrOrNil(item.URL),
		PublishedAt:  time.Unix(item.Time, 0).UTC(),
	}
}
