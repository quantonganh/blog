package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) tagHandler(w http.ResponseWriter, r *http.Request) error {
	tag := mux.Vars(r)["tagName"]
	postsByTag := s.PostService.GetPostsByTag(tag)
	return s.Renderer.RenderPosts(w, r, postsByTag)
}
