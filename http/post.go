package http

import (
	"net/http"
	"path"
	"strings"

	"github.com/gorilla/mux"
)

func (s *Server) postHandler(postsDir string) appHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		uriPath := r.URL.Path
		if hasSuffix(uriPath, []string{"jpg", "jpeg", "png", "gif"}) {
			http.ServeFile(w, r, path.Join(postsDir, uriPath))
		} else {
			if !strings.HasSuffix(uriPath, ".md") {
				uriPath += ".md"
			}
			currentPost := s.PostService.GetPostByURI(uriPath)
			if currentPost == nil {
				return &Error{
					Message: "post not found",
					Status:  http.StatusNotFound,
				}
			}

			relatedPosts := s.PostService.GetRelatedPosts(currentPost)

			previousPost, nextPost := s.PostService.GetPreviousAndNextPost(currentPost)
			if previousPost != nil {
				currentPost.HasPrev = true
			}
			if nextPost != nil {
				currentPost.HasNext = true
			}

			if err := s.Renderer.RenderPost(w, currentPost, relatedPosts, previousPost, nextPost); err != nil {
				return err
			}
		}

		return nil
	}
}

func hasSuffix(path string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}

func (s *Server) postsByDateHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	year := vars["year"]
	month := vars["month"]
	day := vars["day"]

	if err := s.Renderer.RenderPosts(w, r, s.PostService.GetPostsByDate(year, month, day)); err != nil {
		return err
	}

	return nil
}

func (s *Server) postsByMonthHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	year := vars["year"]
	month := vars["month"]

	if err := s.Renderer.RenderPosts(w, r, s.PostService.GetPostsByMonth()[year][month]); err != nil {
		return err
	}

	return nil
}

func (s *Server) postsByYearHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	year := vars["year"]

	if err := s.Renderer.RenderPosts(w, r, s.PostService.GetPostsByYear(year)); err != nil {
		return err
	}

	return nil
}
