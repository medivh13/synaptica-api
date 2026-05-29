package handler

import (
	"log"
	stdhttp "net/http"
	"strconv"
	"strings"

	"synaptica-api/internal/service"
)

const (
	defaultDashboardLimit = 5
	maxDashboardLimit     = 20
)

type DashboardHandler struct {
	service   service.DashboardService
	responder Responder
}

func NewDashboardHandler(service service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		service:   service,
		responder: defaultResponder{},
	}
}

func (h *DashboardHandler) WithResponder(responder Responder) *DashboardHandler {
	h.responder = responder
	return h
}

func (h *DashboardHandler) GetOverview(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	limit, ok := h.parseLimit(w, r)
	if !ok {
		return
	}

	overview, err := h.service.GetOverview(r.Context(), service.DashboardOverviewInput{
		Source: strings.TrimSpace(strings.ToLower(r.URL.Query().Get("source"))),
		Limit:  limit,
	})
	if err != nil {
		log.Printf("get dashboard overview: %v", err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not fetch dashboard overview")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, overview)
}

func (h *DashboardHandler) parseLimit(w stdhttp.ResponseWriter, r *stdhttp.Request) (int, bool) {
	limit := defaultDashboardLimit
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be an integer")
			return 0, false
		}
		if parsedLimit <= 0 {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be greater than 0")
			return 0, false
		}
		limit = parsedLimit
	}
	if limit > maxDashboardLimit {
		limit = maxDashboardLimit
	}
	return limit, true
}
