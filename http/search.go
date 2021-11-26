package http

import (
	"net/http"
)

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return NewError(err, http.StatusBadRequest, "Bad request: invalid search query")
	}

	searchPosts, err := s.PostService.Search(r.FormValue("q"))
	if err != nil {
		return err
	}

	if err := s.Renderer.RenderPosts(w, r, searchPosts); err != nil {
		return err
	}

	return nil
}
