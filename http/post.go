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

func (s *Server) postsByDateHandler(w http.ResponseWriter, r *http.Request) *AppError {
	vars := mux.Vars(r)
	year := vars["year"]
	month := vars["month"]
	day := vars["day"]

	if err := s.Renderer.RenderPosts(w, r, s.PostService.GetPostsByDate(year, month, day)); err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	return nil
}

func (s *Server) postsByMonthHandler(w http.ResponseWriter, r *http.Request) *AppError {
	vars := mux.Vars(r)
	year := vars["year"]
	month := vars["month"]

	if err := s.Renderer.RenderPosts(w, r, s.PostService.GetPostsByMonth(year, month)); err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	return nil
}

func (s *Server) postsByYearHandler(w http.ResponseWriter, r *http.Request) *AppError {
	vars := mux.Vars(r)
	year := vars["year"]

	if err := s.Renderer.RenderPosts(w, r, s.PostService.GetPostsByYear(year)); err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	return nil
}
