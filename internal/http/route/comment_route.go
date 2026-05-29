package route

import (
	"github.com/go-chi/chi/v5"

	"synaptica-api/internal/handler"
)

func RegisterCommentRoutes(r chi.Router, commentHandler *handler.CommentHandler) {
	r.Get("/posts/{post_id}/comments", commentHandler.FindByPostID)
}
