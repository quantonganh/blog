package http

import (
	"net/http"
)

func (s *Server) archivesHandler(w http.ResponseWriter, r *http.Request) *AppError {
	if err := s.Renderer.RenderArchives(w); err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	return nil
}
