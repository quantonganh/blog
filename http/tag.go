package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) tagHandler(w http.ResponseWriter, r *http.Request) *AppError {
	tag := mux.Vars(r)["tagName"]

	postsByTag := s.PostService.GetPostsByTag(tag)

	if err := s.Renderer.RenderPosts(w, r, postsByTag); err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	return nil
}
