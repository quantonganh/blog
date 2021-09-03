package http

import (
	"net/http"
)

func (s *Server) photosHandler(w http.ResponseWriter, r *http.Request) *AppError {
	if err := s.Renderer.RenderPhotos(w); err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	return nil
}
