package handler

import (
	"log"
	stdhttp "net/http"
	"strconv"
	"strings"

	"synaptica-api/internal/repository"
)

const (
	defaultPostLimit = 20
	maxPostLimit     = 100
)

type PostHandler struct {
	postRepo  repository.PostRepository
	responder Responder
}

func NewPostHandler(postRepo repository.PostRepository) *PostHandler {
	return &PostHandler{
		postRepo:  postRepo,
		responder: defaultResponder{},
	}
}

func (h *PostHandler) WithResponder(responder Responder) *PostHandler {
	h.responder = responder
	return h
}

func (h *PostHandler) FindRecent(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	filter, ok := h.parseFilter(w, r)
	if !ok {
		return
	}

	posts, err := h.postRepo.FindRecent(r.Context(), filter)
	if err != nil {
		log.Printf("find recent posts: %v", err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not fetch posts")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, posts)
}

func (h *PostHandler) parseFilter(w stdhttp.ResponseWriter, r *stdhttp.Request) (repository.PostFilter, bool) {
	source := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("source")))
	if source != "" && source != "hackernews" && source != "reddit" {
		h.responder.WriteError(w, stdhttp.StatusBadRequest, "source must be hackernews or reddit")
		return repository.PostFilter{}, false
	}

	limit := defaultPostLimit
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be an integer")
			return repository.PostFilter{}, false
		}
		if parsedLimit <= 0 {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be greater than 0")
			return repository.PostFilter{}, false
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
			return repository.PostFilter{}, false
		}
		if parsedOffset < 0 {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "offset must be greater than or equal to 0")
			return repository.PostFilter{}, false
		}
		offset = parsedOffset
	}

	return repository.PostFilter{
		Source: source,
		Query:  strings.TrimSpace(r.URL.Query().Get("q")),
		Limit:  limit,
		Offset: offset,
	}, true
}
