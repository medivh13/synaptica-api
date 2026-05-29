package handler

import (
	"errors"
	"log"
	stdhttp "net/http"
	"strconv"

	"synaptica-api/internal/service"
)

type HackerNewsIngestionHandler struct {
	service   service.HackerNewsIngestionService
	responder Responder
}

func NewHackerNewsIngestionHandler(service service.HackerNewsIngestionService) *HackerNewsIngestionHandler {
	return &HackerNewsIngestionHandler{
		service:   service,
		responder: defaultResponder{},
	}
}

func (h *HackerNewsIngestionHandler) WithResponder(responder Responder) *HackerNewsIngestionHandler {
	h.responder = responder
	return h
}

func (h *HackerNewsIngestionHandler) IngestStories(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	feed := r.URL.Query().Get("feed")

	limit := 0
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be an integer")
			return
		}
		limit = parsedLimit
	}

	result, err := h.service.IngestStories(r.Context(), feed, limit)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "invalid hacker news ingestion request")
		case errors.Is(err, service.ErrNotFound):
			h.responder.WriteError(w, stdhttp.StatusNotFound, "resource not found")
		default:
			log.Printf("ingest hacker news stories: %v", err)
			h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not ingest hacker news stories")
		}
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, result)
}
