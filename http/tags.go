package http

import (
	"net/http"
)

func (s *Server) tagsHandler(w http.ResponseWriter, r *http.Request) error {
	return s.Renderer.RenderTags(w)
}
