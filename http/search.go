package http

import (
	"net/http"
)

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) *AppError {
	if err := r.ParseForm(); err != nil {
		return &AppError{
			Error:   err,
			Message: "failed to parse raw query",
			Code:    http.StatusInternalServerError,
		}
	}

	searchPosts, err := s.PostService.Search(r.FormValue("q"))
	if err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	if err := s.Renderer.RenderPosts(w, r, searchPosts); err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	return nil
}
