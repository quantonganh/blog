package http

import (
	"net/http"
)

func (s *Server) aboutHandler(w http.ResponseWriter, r *http.Request) *AppError {
	currentPost := s.PostService.GetPostByURI("2019/09/19/about")

	relatedPosts := s.PostService.GetRelatedPosts(currentPost)

	previousPost, nextPost := s.PostService.GetPreviousAndNextPost(currentPost)
	if previousPost != nil {
		currentPost.HasPrev = true
	}
	if nextPost != nil {
		currentPost.HasNext = true
	}

	if err := s.Renderer.RenderPost(w, currentPost, relatedPosts, previousPost, nextPost); err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	return nil
}
