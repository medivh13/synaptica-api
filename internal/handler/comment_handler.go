package handler

import (
	"log"
	stdhttp "net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"synaptica-api/internal/repository"
)

const (
	defaultCommentLimit = 50
	maxCommentLimit     = 200
)

type CommentHandler struct {
	commentRepo repository.CommentRepository
	responder   Responder
}

func NewCommentHandler(commentRepo repository.CommentRepository) *CommentHandler {
	return &CommentHandler{
		commentRepo: commentRepo,
		responder:   defaultResponder{},
	}
}

func (h *CommentHandler) WithResponder(responder Responder) *CommentHandler {
	h.responder = responder
	return h
}

func (h *CommentHandler) FindByPostID(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	filter, ok := h.parseFilter(w, r)
	if !ok {
		return
	}

	comments, err := h.commentRepo.FindByPostID(r.Context(), filter)
	if err != nil {
		log.Printf("find comments by post id: %v", err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not fetch comments")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, comments)
}

func (h *CommentHandler) parseFilter(w stdhttp.ResponseWriter, r *stdhttp.Request) (repository.CommentFilter, bool) {
	postID := strings.TrimSpace(chi.URLParam(r, "post_id"))
	if postID == "" {
		h.responder.WriteError(w, stdhttp.StatusBadRequest, "post_id is required")
		return repository.CommentFilter{}, false
	}

	limit := defaultCommentLimit
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be an integer")
			return repository.CommentFilter{}, false
		}
		if parsedLimit <= 0 {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "limit must be greater than 0")
			return repository.CommentFilter{}, false
		}
		limit = parsedLimit
	}
	if limit > maxCommentLimit {
		limit = maxCommentLimit
	}

	offset := 0
	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		parsedOffset, err := strconv.Atoi(rawOffset)
		if err != nil {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "offset must be an integer")
			return repository.CommentFilter{}, false
		}
		if parsedOffset < 0 {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "offset must be greater than or equal to 0")
			return repository.CommentFilter{}, false
		}
		offset = parsedOffset
	}

	var depth *int
	if rawDepth := r.URL.Query().Get("depth"); rawDepth != "" {
		parsedDepth, err := strconv.Atoi(rawDepth)
		if err != nil {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "depth must be an integer")
			return repository.CommentFilter{}, false
		}
		if parsedDepth < 0 {
			h.responder.WriteError(w, stdhttp.StatusBadRequest, "depth must be greater than or equal to 0")
			return repository.CommentFilter{}, false
		}
		depth = &parsedDepth
	}

	return repository.CommentFilter{
		PostID: postID,
		Limit:  limit,
		Offset: offset,
		Depth:  depth,
	}, true
}
