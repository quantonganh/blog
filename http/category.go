package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) categoryHandler(w http.ResponseWriter, r *http.Request) error {
	category := mux.Vars(r)["categoryName"]

	postsByCategory := s.PostService.GetPostsByCategory(category)

	if err := s.Renderer.RenderPosts(w, r, postsByCategory); err != nil {
		return err
	}

	return nil
}
