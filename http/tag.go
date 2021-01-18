package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) tagHandler(w http.ResponseWriter, r *http.Request) *AppError {
	tag := mux.Vars(r)["tagName"]

	postsByTag, err := s.PostService.GetPostsByTag(tag)
	if err != nil {
		return &AppError{
			Error:   err,
			Message: fmt.Sprintf("failed to get posts by tag: %s", tag),
			Code:    http.StatusInternalServerError,
		}
	}

	if err := s.Renderer.RenderPosts(w, r, postsByTag); err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	return nil
}
