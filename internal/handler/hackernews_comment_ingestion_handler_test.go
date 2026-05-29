package handler

import (
	"context"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"synaptica-api/internal/service"
)

type fakeHNCommentIngestionService struct {
	result *service.HNCommentIngestionResult
	err    error
}

func (s fakeHNCommentIngestionService) IngestPostComments(ctx context.Context, input service.HNCommentIngestionInput) (*service.HNCommentIngestionResult, error) {
	return s.result, s.err
}

func TestHackerNewsCommentIngestionHandlerSuccessResponse(t *testing.T) {
	handler := NewHackerNewsCommentIngestionHandler(fakeHNCommentIngestionService{
		result: &service.HNCommentIngestionResult{
			PostID:     "post-1",
			ExternalID: "48192224",
			Fetched:    10,
			Inserted:   8,
			Skipped:    2,
			MaxDepth:   3,
		},
	})

	req := httptest.NewRequest(stdhttp.MethodPost, "/api/v1/ingestion/hackernews/posts/post-1/comments?max_depth=3&limit=10", nil)
	rec := httptest.NewRecorder()

	handler.IngestPostComments(rec, req)

	if rec.Code != stdhttp.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected application/json content type, got %q", contentType)
	}

	var body struct {
		Data service.HNCommentIngestionResult `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Data.PostID != "post-1" ||
		body.Data.ExternalID != "48192224" ||
		body.Data.Fetched != 10 ||
		body.Data.Inserted != 8 ||
		body.Data.Skipped != 2 ||
		body.Data.MaxDepth != 3 {
		t.Fatalf("unexpected response body: %+v", body.Data)
	}
}

func TestHackerNewsCommentIngestionHandlerDebugResponse(t *testing.T) {
	handler := NewHackerNewsCommentIngestionHandler(fakeHNCommentIngestionService{})

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/v1/ingestion/hackernews/debug-response", nil)
	rec := httptest.NewRecorder()

	handler.DebugResponse(rec, req)

	if rec.Code != stdhttp.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.Status != "ok" {
		t.Fatalf("expected status ok, got %q", body.Data.Status)
	}
}
