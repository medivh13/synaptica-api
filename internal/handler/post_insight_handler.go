package handler

import (
	"errors"
	"log"
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"

	"github.com/medivh13/synaptica-api/internal/service"
)

type PostInsightHandler struct {
	service   service.PostInsightService
	responder Responder
}

func NewPostInsightHandler(service service.PostInsightService) *PostInsightHandler {
	return &PostInsightHandler{
		service:   service,
		responder: defaultResponder{},
	}
}

func (h *PostInsightHandler) WithResponder(responder Responder) *PostInsightHandler {
	h.responder = responder
	return h
}

func (h *PostInsightHandler) GetPostInsight(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	insight, err := h.service.GetPostInsight(r.Context(), chi.URLParam(r, "post_id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "invalid post insight request")
		case errors.Is(err, service.ErrNotFound):
			h.responder.WriteError(w, stdhttp.StatusNotFound, "post not found")
		default:
			log.Printf("get post insight: %v", err)
			h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not fetch post insight")
		}
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, insight)
}
