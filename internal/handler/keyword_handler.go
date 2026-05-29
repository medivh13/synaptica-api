package handler

import (
	"encoding/json"
	"errors"
	"log"
	stdhttp "net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"synaptica-api/internal/service"
)

type KeywordHandler struct {
	service   service.KeywordService
	responder Responder
}

type createKeywordRequest struct {
	Keyword     string `json:"keyword"`
	Description string `json:"description"`
}

func NewKeywordHandler(service service.KeywordService) *KeywordHandler {
	return &KeywordHandler{
		service:   service,
		responder: defaultResponder{},
	}
}

func (h *KeywordHandler) WithResponder(responder Responder) *KeywordHandler {
	h.responder = responder
	return h
}

func (h *KeywordHandler) Create(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req createKeywordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responder.WriteError(w, stdhttp.StatusBadRequest, "invalid JSON request body")
		return
	}

	keyword, err := h.service.Create(r.Context(), service.KeywordCreateInput{
		Keyword:     req.Keyword,
		Description: req.Description,
	})
	if err != nil {
		h.writeServiceError(w, err, "could not create keyword")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusCreated, keyword)
}

func (h *KeywordHandler) FindAllActive(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	keywords, err := h.service.FindAllActive(r.Context())
	if err != nil {
		log.Printf("find active keywords: %v", err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not fetch keywords")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, keywords)
}

func (h *KeywordHandler) MatchPosts(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	result, err := h.service.MatchPosts(r.Context(), chi.URLParam(r, "keyword_id"))
	if err != nil {
		h.writeServiceError(w, err, "could not match keyword posts")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, result)
}

func (h *KeywordHandler) FindPosts(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	limit, offset, ok := h.parsePagination(w, r)
	if !ok {
		return
	}

	posts, err := h.service.FindPosts(r.Context(), chi.URLParam(r, "keyword_id"), limit, offset)
	if err != nil {
		h.writeServiceError(w, err, "could not fetch keyword posts")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, posts)
}

func (h *KeywordHandler) parsePagination(w stdhttp.ResponseWriter, r *stdhttp.Request) (int, int, bool) {
	limit := defaultPostLimit
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be an integer")
			return 0, 0, false
		}
		if parsedLimit <= 0 {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be greater than 0")
			return 0, 0, false
		}
		limit = parsedLimit
	}
	if limit > maxPostLimit {
		limit = maxPostLimit
	}

	offset := 0
	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		parsedOffset, err := strconv.Atoi(rawOffset)
		if err != nil {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "offset must be an integer")
			return 0, 0, false
		}
		if parsedOffset < 0 {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "offset must be greater than or equal to 0")
			return 0, 0, false
		}
		offset = parsedOffset
	}

	return limit, offset, true
}

func (h *KeywordHandler) writeServiceError(w stdhttp.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		h.responder.WriteError(w, stdhttp.StatusBadRequest, "invalid keyword request")
	case errors.Is(err, service.ErrNotFound):
		h.responder.WriteError(w, stdhttp.StatusNotFound, "keyword not found")
	default:
		log.Printf("%s: %v", fallback, err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, fallback)
	}
}
