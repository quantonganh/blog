package http

import (
	"net/http"
)

func (s *Server) tagsHandler(w http.ResponseWriter, r *http.Request) *AppError {
	if err := s.Renderer.RenderTags(w); err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	return nil
}
