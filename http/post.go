package http

import (
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

func (s *Server) postHandler(w http.ResponseWriter, r *http.Request) *AppError {
	vars := mux.Vars(r)
	year := vars["year"]
	month := vars["month"]
	day := vars["day"]
	fileName := vars["postName"]

	currentPost := s.PostService.GetPostByURI(filepath.Join(year, month, day, fileName))

	relatedPosts, err := s.PostService.GetRelatedPosts(currentPost)
	if err != nil {
		return &AppError{
			Error:   err,
			Message: "failed to get related posts",
			Code:    http.StatusInternalServerError,
		}
	}

	previousPost, nextPost := s.PostService.GetPreviousAndNextPost(currentPost)
	if previousPost != nil {
		currentPost.HasPrev = true
	}
	if nextPost != nil {
		currentPost.HasNext = true
	}

	if err := s.Renderer.RenderPost(w, currentPost, relatedPosts, previousPost, currentPost); err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	return nil
}
