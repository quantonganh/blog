package http

import (
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

func (s *Server) postHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	year := vars["year"]
	month := vars["month"]
	day := vars["day"]
	fileName := vars["postName"]

	currentPost := s.PostService.GetPostByURI(filepath.Join(year, month, day, fileName))

	relatedPosts, err := s.PostService.GetRelatedPosts(currentPost)
	if err != nil {
		return err
	}

	previousPost, nextPost := s.PostService.GetPreviousAndNextPost(currentPost)
	if previousPost != nil {
		currentPost.HasPrev = true
	}
	if nextPost != nil {
		currentPost.HasNext = true
	}

	return s.Renderer.RenderPost(w, currentPost, relatedPosts, previousPost, currentPost)
}
