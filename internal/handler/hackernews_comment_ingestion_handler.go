package handler

import (
	"errors"
	"log"
	stdhttp "net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"synaptica-api/internal/service"
)

type HackerNewsCommentIngestionHandler struct {
	service   service.HackerNewsCommentIngestionService
	responder Responder
}

func NewHackerNewsCommentIngestionHandler(service service.HackerNewsCommentIngestionService) *HackerNewsCommentIngestionHandler {
	return &HackerNewsCommentIngestionHandler{
		service:   service,
		responder: defaultResponder{},
	}
}

func (h *HackerNewsCommentIngestionHandler) WithResponder(responder Responder) *HackerNewsCommentIngestionHandler {
	h.responder = responder
	return h
}

func (h *HackerNewsCommentIngestionHandler) IngestPostComments(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	maxDepth, ok := h.parsePositiveInt(w, r, "max_depth")
	if !ok {
		return
	}
	limit, ok := h.parsePositiveInt(w, r, "limit")
	if !ok {
		return
	}

	result, err := h.service.IngestPostComments(r.Context(), service.HNCommentIngestionInput{
		PostID:   chi.URLParam(r, "post_id"),
		MaxDepth: maxDepth,
		Limit:    limit,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "invalid hacker news comment ingestion request")
		case errors.Is(err, service.ErrNotFound):
			h.responder.WriteError(w, stdhttp.StatusNotFound, "resource not found")
		default:
			log.Printf("ingest hacker news comments: %v", err)
			h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not ingest hacker news comments")
		}
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, result)
	return
}

func (h *HackerNewsCommentIngestionHandler) DebugResponse(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	h.responder.WriteSuccess(w, stdhttp.StatusOK, map[string]string{"status": "ok"})
	return
}

func (h *HackerNewsCommentIngestionHandler) parsePositiveInt(w stdhttp.ResponseWriter, r *stdhttp.Request, name string) (int, bool) {
	rawValue := r.URL.Query().Get(name)
	if rawValue == "" {
		return 0, true
	}

	value, err := strconv.Atoi(rawValue)
	if err != nil {
		h.responder.WriteError(w, stdhttp.StatusBadRequest, name+" must be an integer")
		return 0, false
	}
	if value <= 0 {
		h.responder.WriteError(w, stdhttp.StatusBadRequest, name+" must be greater than 0")
		return 0, false
	}

	return value, true
}
