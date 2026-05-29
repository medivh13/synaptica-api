package handler

import (
	"errors"
	"log"
	stdhttp "net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/medivh13/synaptica-api/internal/service"
)

type RedditIngestionHandler struct {
	service   service.RedditIngestionService
	responder Responder
}

func NewRedditIngestionHandler(service service.RedditIngestionService) *RedditIngestionHandler {
	return &RedditIngestionHandler{
		service:   service,
		responder: defaultResponder{},
	}
}

func (h *RedditIngestionHandler) WithResponder(responder Responder) *RedditIngestionHandler {
	h.responder = responder
	return h
}

func (h *RedditIngestionHandler) IngestLatestPosts(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	subreddit := chi.URLParam(r, "subreddit")

	limit := 0
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be an integer")
			return
		}
		limit = parsedLimit
	}

	result, err := h.service.IngestLatestPosts(r.Context(), subreddit, limit)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "invalid ingestion request")
		case errors.Is(err, service.ErrNotFound):
			h.responder.WriteError(w, stdhttp.StatusNotFound, "resource not found")
		default:
			log.Printf("ingest latest reddit posts: %v", err)
			h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not ingest reddit posts")
		}
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, result)
}
