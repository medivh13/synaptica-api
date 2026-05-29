package handler

import (
	"log"
	stdhttp "net/http"

	"github.com/medivh13/synaptica-api/internal/repository"
)

type SubredditHandler struct {
	subredditRepo repository.SubredditRepository
	responder     Responder
}

func NewSubredditHandler(subredditRepo repository.SubredditRepository) *SubredditHandler {
	return &SubredditHandler{
		subredditRepo: subredditRepo,
		responder:     defaultResponder{},
	}
}

func (h *SubredditHandler) WithResponder(responder Responder) *SubredditHandler {
	h.responder = responder
	return h
}

func (h *SubredditHandler) FindAll(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	subreddits, err := h.subredditRepo.FindAll(r.Context())
	if err != nil {
		log.Printf("find all subreddits: %v", err)
		h.responder.WriteError(w, stdhttp.StatusInternalServerError, "could not fetch subreddits")
		return
	}

	h.responder.WriteSuccess(w, stdhttp.StatusOK, subreddits)
}
