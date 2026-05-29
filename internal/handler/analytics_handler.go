package handler

import (
	"log"
	stdhttp "net/http"
	"strconv"
	"strings"

	"github.com/medivh13/synaptica-api/internal/service"
)

const (
	defaultAnalyticsLimit = 24
	maxAnalyticsLimit     = 168
	defaultRankingLimit   = 10
	maxRankingLimit       = 50
)

type AnalyticsHandler struct {
	service   service.AnalyticsService
	responder Responder
}

func NewAnalyticsHandler(service service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		service:   service,
		responder: defaultResponder{},
	}
}

func (h *AnalyticsHandler) WithResponder(responder Responder) *AnalyticsHandler {
	h.responder = responder
	return h
}

func (h *AnalyticsHandler) GetEmotionTimeline(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	input, ok := h.parseInput(w, r, true)
	if !ok {
		return
	}

	buckets, err := h.service.GetEmotionTimeline(r.Context(), input)
	if err != nil {
		log.Printf("get emotion timeline: %v", err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not fetch emotion timeline")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, buckets)
}

func (h *AnalyticsHandler) GetEmotionSummary(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	input, ok := h.parseInput(w, r, false)
	if !ok {
		return
	}

	summary, err := h.service.GetEmotionSummary(r.Context(), input)
	if err != nil {
		log.Printf("get emotion summary: %v", err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not fetch emotion summary")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, summary)
}

func (h *AnalyticsHandler) GetTrendingKeywords(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	input, ok := h.parseRankingInput(w, r, true)
	if !ok {
		return
	}

	keywords, err := h.service.GetTrendingKeywords(r.Context(), input)
	if err != nil {
		log.Printf("get trending keywords: %v", err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not fetch trending keywords")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, keywords)
}

func (h *AnalyticsHandler) GetHotTopics(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	input, ok := h.parseRankingInput(w, r, false)
	if !ok {
		return
	}

	topics, err := h.service.GetHotTopics(r.Context(), input)
	if err != nil {
		log.Printf("get hot topics: %v", err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not fetch hot topics")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, topics)
}

func (h *AnalyticsHandler) parseInput(w stdhttp.ResponseWriter, r *stdhttp.Request, includeInterval bool) (service.AnalyticsInput, bool) {
	entityType := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("entity_type")))
	if entityType == "" {
		entityType = "comment"
	}
	if entityType != "post" && entityType != "comment" && entityType != "all" {
		h.responder.WriteError(w, stdhttp.StatusBadRequest, "entity_type must be post, comment, or all")
		return service.AnalyticsInput{}, false
	}

	interval := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("interval")))
	if interval == "" {
		interval = "hour"
	}
	if includeInterval && interval != "hour" && interval != "day" {
		h.responder.WriteError(w, stdhttp.StatusBadRequest, "interval must be hour or day")
		return service.AnalyticsInput{}, false
	}

	limit := defaultAnalyticsLimit
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be an integer")
			return service.AnalyticsInput{}, false
		}
		if parsedLimit <= 0 {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be greater than 0")
			return service.AnalyticsInput{}, false
		}
		limit = parsedLimit
	}
	if limit > maxAnalyticsLimit {
		limit = maxAnalyticsLimit
	}

	return service.AnalyticsInput{
		EntityType: entityType,
		Source:     strings.TrimSpace(strings.ToLower(r.URL.Query().Get("source"))),
		PostID:     strings.TrimSpace(r.URL.Query().Get("post_id")),
		KeywordID:  strings.TrimSpace(r.URL.Query().Get("keyword_id")),
		Interval:   interval,
		Limit:      limit,
	}, true
}

func (h *AnalyticsHandler) parseRankingInput(w stdhttp.ResponseWriter, r *stdhttp.Request, includeEntityType bool) (service.AnalyticsInput, bool) {
	entityType := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("entity_type")))
	if entityType == "" {
		entityType = "all"
	}
	if includeEntityType && entityType != "post" && entityType != "comment" && entityType != "all" {
		h.responder.WriteError(w, stdhttp.StatusBadRequest, "entity_type must be post, comment, or all")
		return service.AnalyticsInput{}, false
	}

	limit := defaultRankingLimit
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be an integer")
			return service.AnalyticsInput{}, false
		}
		if parsedLimit <= 0 {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be greater than 0")
			return service.AnalyticsInput{}, false
		}
		limit = parsedLimit
	}
	if limit > maxRankingLimit {
		limit = maxRankingLimit
	}

	return service.AnalyticsInput{
		EntityType: entityType,
		Source:     strings.TrimSpace(strings.ToLower(r.URL.Query().Get("source"))),
		Limit:      limit,
	}, true
}
