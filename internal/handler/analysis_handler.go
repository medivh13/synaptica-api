package handler

import (
	"log"
	stdhttp "net/http"
	"strconv"
	"strings"

	"synaptica-api/internal/service"
)

const (
	defaultAnalysisRunLimit    = 50
	maxAnalysisRunLimit        = 200
	defaultAnalysisResultLimit = 20
	maxAnalysisResultLimit     = 100
)

type AnalysisHandler struct {
	service   service.PostAnalysisService
	responder Responder
}

func NewAnalysisHandler(service service.PostAnalysisService) *AnalysisHandler {
	return &AnalysisHandler{
		service:   service,
		responder: defaultResponder{},
	}
}

func (h *AnalysisHandler) WithResponder(responder Responder) *AnalysisHandler {
	h.responder = responder
	return h
}

func (h *AnalysisHandler) RunPostAnalysis(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	limit, ok := h.parseLimit(w, r, defaultAnalysisRunLimit, maxAnalysisRunLimit)
	if !ok {
		return
	}

	result, err := h.service.RunPostAnalysis(r.Context(), service.RunPostAnalysisInput{
		Source:    strings.TrimSpace(strings.ToLower(r.URL.Query().Get("source"))),
		KeywordID: strings.TrimSpace(r.URL.Query().Get("keyword_id")),
		Limit:     limit,
	})
	if err != nil {
		log.Printf("run post analysis: %v", err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not run post analysis")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, result)
}

func (h *AnalysisHandler) FindPostAnalysisResults(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	limit, ok := h.parseLimit(w, r, defaultAnalysisResultLimit, maxAnalysisResultLimit)
	if !ok {
		return
	}
	offset, ok := h.parseOffset(w, r)
	if !ok {
		return
	}

	posts, err := h.service.FindAnalyzedPosts(r.Context(), service.FindAnalyzedPostsInput{
		Source:    strings.TrimSpace(strings.ToLower(r.URL.Query().Get("source"))),
		KeywordID: strings.TrimSpace(r.URL.Query().Get("keyword_id")),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		log.Printf("find post analysis results: %v", err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not fetch post analysis results")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, posts)
}

func (h *AnalysisHandler) parseLimit(w stdhttp.ResponseWriter, r *stdhttp.Request, defaultLimit int, maxLimit int) (int, bool) {
	limit := defaultLimit
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
	if limit > maxLimit {
		limit = maxLimit
	}
	return limit, true
}

func (h *AnalysisHandler) parseOffset(w stdhttp.ResponseWriter, r *stdhttp.Request) (int, bool) {
	offset := 0
	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		parsedOffset, err := strconv.Atoi(rawOffset)
		if err != nil {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "offset must be an integer")
			return 0, false
		}
		if parsedOffset < 0 {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "offset must be greater than or equal to 0")
			return 0, false
		}
		offset = parsedOffset
	}
	return offset, true
}
