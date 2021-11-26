package http

import (
	"net/http"
)

func (s *Server) archivesHandler(w http.ResponseWriter, r *http.Request) error {
	return s.Renderer.RenderArchives(w)
}
