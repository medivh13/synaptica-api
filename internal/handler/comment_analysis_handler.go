package handler

import (
	"log"
	stdhttp "net/http"
	"strconv"
	"strings"

	"synaptica-api/internal/service"
)

const (
	defaultCommentAnalysisRunLimit    = 100
	maxCommentAnalysisRunLimit        = 500
	defaultCommentAnalysisResultLimit = 50
	maxCommentAnalysisResultLimit     = 200
)

type CommentAnalysisHandler struct {
	service   service.CommentAnalysisService
	responder Responder
}

func NewCommentAnalysisHandler(service service.CommentAnalysisService) *CommentAnalysisHandler {
	return &CommentAnalysisHandler{
		service:   service,
		responder: defaultResponder{},
	}
}

func (h *CommentAnalysisHandler) WithResponder(responder Responder) *CommentAnalysisHandler {
	h.responder = responder
	return h
}

func (h *CommentAnalysisHandler) RunCommentAnalysis(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	limit, ok := h.parseLimit(w, r, defaultCommentAnalysisRunLimit, maxCommentAnalysisRunLimit)
	if !ok {
		return
	}

	result, err := h.service.RunCommentAnalysis(r.Context(), service.RunCommentAnalysisInput{
		PostID:    strings.TrimSpace(r.URL.Query().Get("post_id")),
		Source:    strings.TrimSpace(strings.ToLower(r.URL.Query().Get("source"))),
		KeywordID: strings.TrimSpace(r.URL.Query().Get("keyword_id")),
		Limit:     limit,
	})
	if err != nil {
		log.Printf("run comment analysis: %v", err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not run comment analysis")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, result)
}

func (h *CommentAnalysisHandler) FindCommentAnalysisResults(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	limit, ok := h.parseLimit(w, r, defaultCommentAnalysisResultLimit, maxCommentAnalysisResultLimit)
	if !ok {
		return
	}
	offset, ok := h.parseOffset(w, r)
	if !ok {
		return
	}

	comments, err := h.service.FindAnalyzedComments(r.Context(), service.FindAnalyzedCommentsInput{
		PostID:    strings.TrimSpace(r.URL.Query().Get("post_id")),
		Source:    strings.TrimSpace(strings.ToLower(r.URL.Query().Get("source"))),
		KeywordID: strings.TrimSpace(r.URL.Query().Get("keyword_id")),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		log.Printf("find comment analysis results: %v", err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not fetch comment analysis results")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, comments)
}

func (h *CommentAnalysisHandler) parseLimit(w stdhttp.ResponseWriter, r *stdhttp.Request, defaultLimit int, maxLimit int) (int, bool) {
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

func (h *CommentAnalysisHandler) parseOffset(w stdhttp.ResponseWriter, r *stdhttp.Request) (int, bool) {
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
