package http

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/quantonganh/blog/http/html"
)

func (s *Server) tagHandler(w http.ResponseWriter, r *http.Request) error {
	tag := mux.Vars(r)["tagName"]

	postsByTag, err := s.PostService.GetPostsByTag(tag)
	if err != nil {
		return err
	}

	if err := html.NewPost(s.Templates).Render(w, r, postsByTag); err != nil {
		return err
	}

	return nil
}
