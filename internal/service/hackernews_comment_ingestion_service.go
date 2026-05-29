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

type HackerNewsCommentIngestionService interface {
	IngestPostComments(ctx context.Context, input HNCommentIngestionInput) (*HNCommentIngestionResult, error)
}

type HNCommentIngestionInput struct {
	PostID   string
	MaxDepth int
	Limit    int
}

type HNCommentIngestionResult struct {
	PostID     string `json:"post_id"`
	ExternalID string `json:"external_id"`
	Fetched    int    `json:"fetched"`
	Inserted   int    `json:"inserted"`
	Skipped    int    `json:"skipped"`
	MaxDepth   int    `json:"max_depth"`
}

type hackerNewsCommentIngestionService struct {
	hnClient       hackernews.Client
	dataSourceRepo repository.DataSourceRepository
	postRepo       repository.PostRepository
	commentRepo    repository.CommentRepository
}

func NewHackerNewsCommentIngestionService(
	hnClient hackernews.Client,
	dataSourceRepo repository.DataSourceRepository,
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
) HackerNewsCommentIngestionService {
	return &hackerNewsCommentIngestionService{
		hnClient:       hnClient,
		dataSourceRepo: dataSourceRepo,
		postRepo:       postRepo,
		commentRepo:    commentRepo,
	}
}

func (s *hackerNewsCommentIngestionService) IngestPostComments(ctx context.Context, input HNCommentIngestionInput) (*HNCommentIngestionResult, error) {
	postID := strings.TrimSpace(input.PostID)
	if postID == "" {
		return nil, fmt.Errorf("%w: post id is required", ErrInvalidInput)
	}

	maxDepth := input.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 3
	}
	if maxDepth > 5 {
		maxDepth = 5
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 300 {
		limit = 300
	}

	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: post %q", ErrNotFound, postID)
		}
		return nil, fmt.Errorf("find post: %w", err)
	}

	hnDataSource, err := s.dataSourceRepo.FindByType(ctx, "hackernews")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: hackernews data source", ErrNotFound)
		}
		return nil, fmt.Errorf("find hackernews data source: %w", err)
	}
	if post.SourceID != hnDataSource.ID {
		return nil, fmt.Errorf("%w: post is not from hackernews", ErrInvalidInput)
	}

	rootID, err := strconv.ParseInt(post.ExternalID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: post external id is not a hackernews item id", ErrInvalidInput)
	}

	rootItem, err := s.hnClient.FetchItem(ctx, rootID)
	if err != nil {
		return nil, fmt.Errorf("fetch hackernews story item: %w", err)
	}
	if rootItem == nil {
		return nil, fmt.Errorf("%w: hackernews story item", ErrNotFound)
	}

	comments := make([]model.Comment, 0, min(limit, len(rootItem.Kids)))
	for _, kidID := range rootItem.Kids {
		if len(comments) >= limit {
			break
		}
		if err := s.collectComments(ctx, post, kidID, 1, maxDepth, limit, &comments); err != nil {
			return nil, err
		}
	}

	upsertResult, err := s.commentRepo.UpsertMany(ctx, comments)
	if err != nil {
		return nil, fmt.Errorf("store hackernews comments: %w", err)
	}

	return &HNCommentIngestionResult{
		PostID:     post.ID,
		ExternalID: post.ExternalID,
		Fetched:    len(comments),
		Inserted:   upsertResult.Inserted,
		Skipped:    upsertResult.Skipped,
		MaxDepth:   maxDepth,
	}, nil
}

func (s *hackerNewsCommentIngestionService) collectComments(
	ctx context.Context,
	post *model.Post,
	commentID int64,
	depth int,
	maxDepth int,
	limit int,
	comments *[]model.Comment,
) error {
	if depth > maxDepth || len(*comments) >= limit {
		return nil
	}

	item, err := s.hnClient.FetchItem(ctx, commentID)
	if err != nil {
		return fmt.Errorf("fetch hackernews comment %d: %w", commentID, err)
	}
	if shouldSkipHNComment(item) {
		return nil
	}

	*comments = append(*comments, mapHNComment(post, item, depth))
	if len(*comments) >= limit {
		return nil
	}

	for _, kidID := range item.Kids {
		if len(*comments) >= limit {
			break
		}
		if err := s.collectComments(ctx, post, kidID, depth+1, maxDepth, limit, comments); err != nil {
			return err
		}
	}

	return nil
}

func shouldSkipHNComment(item *hackernews.Item) bool {
	return item == nil ||
		item.Deleted ||
		item.Dead ||
		item.Type != "comment" ||
		strings.TrimSpace(item.Text) == ""
}

func mapHNComment(post *model.Post, item *hackernews.Item, depth int) model.Comment {
	parentExternalID := strconv.FormatInt(item.Parent, 10)
	return model.Comment{
		SourceID:         post.SourceID,
		SubredditID:      post.SubredditID,
		PostID:           post.ID,
		ExternalID:       strconv.FormatInt(item.ID, 10),
		ParentExternalID: &parentExternalID,
		ParentCommentID:  nil,
		AuthorName:       stringPtrOrNil(item.By),
		Body:             item.Text,
		Score:            0,
		Depth:            depth,
		PublishedAt:      time.Unix(item.Time, 0).UTC(),
	}
}
