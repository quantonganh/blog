package http

import (
	"net/http"
	"os"

	"github.com/blevesearch/bleve"
)

func (s *Server) searchHandler(indexPath string) ErrHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) *AppError {
		var (
			index bleve.Index
			err   error
		)
		if _, err = os.Stat(indexPath); os.IsNotExist(err) {
			index, err = s.PostService.IndexPosts(indexPath)
			if err != nil {
				return &AppError{
					Error:   err,
					Message: "failed to index posts",
					Code:    http.StatusInternalServerError,
				}
			}
		} else if err == nil {
			index, err = bleve.OpenUsing(indexPath, map[string]interface{}{
				"read_only": true,
			})
			if err != nil {
				return &AppError{
					Error:   err,
					Message: "failed to open index",
					Code:    http.StatusInternalServerError,
				}
			}
		}
		defer func() {
			_ = index.Close()
		}()

		if err := r.ParseForm(); err != nil {
			return &AppError{
				Error:   err,
				Message: "failed to parse raw query",
				Code:    http.StatusInternalServerError,
			}
		}

		searchPosts, err := s.PostService.Search(index, r.FormValue("q"))
		if err != nil {
			return &AppError{
				Error: err,
				Code:  http.StatusInternalServerError,
			}
		}

		if err := s.Renderer.RenderPosts(w, r, searchPosts); err != nil {
			return &AppError{
				Error: err,
				Code:  http.StatusInternalServerError,
			}
		}

		return nil
	}
}
