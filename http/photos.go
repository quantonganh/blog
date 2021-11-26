package http

import (
	"net/http"
)

func (s *Server) photosHandler(w http.ResponseWriter, r *http.Request) error {
	return s.Renderer.RenderPhotos(w)
}
