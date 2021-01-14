package http

import (
	"net/http"
	"os"

	"github.com/blevesearch/bleve"
	"github.com/pkg/errors"

	"github.com/quantonganh/blog/http/html"
)

const (
	indexPath = "posts.bleve"
)

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) error {
	var (
		index bleve.Index
		err   error
	)
	if _, err = os.Stat(indexPath); os.IsNotExist(err) {
		index, err = s.PostService.IndexPosts(indexPath)
		if err != nil {
			return errors.Errorf("failed to index posts: %v", err)
		}
	} else if err == nil {
		index, err = bleve.OpenUsing(indexPath, map[string]interface{}{
			"read_only": true,
		})
		if err != nil {
			return errors.Errorf("failed to open index at %s: %v", indexPath, err)
		}
	}
	defer func() {
		_ = index.Close()
	}()

	if err := r.ParseForm(); err != nil {
		return errors.Errorf("failed to parse form: %v", err)
	}

	searchPosts, err := s.PostService.Search(index, r.FormValue("q"))
	if err != nil {
		return errors.Errorf("failed to search: %v", err)
	}

	if err := html.NewPost(s.Templates).Render(w, r, searchPosts); err != nil {
		return errors.Errorf("failed to render HTML: %v", err)
	}

	return nil
}
