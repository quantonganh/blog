package http

import (
	"net/http"
	"path/filepath"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"

	"github.com/quantonganh/blog"
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

	remark42, err := blog.GetRemarkURL()
	if err != nil {
		return err
	}

	data := pongo2.Context{"title": currentPost.Title, "currentPost": currentPost, "relatedPosts": relatedPosts, "previousPost": previousPost, "nextPost": nextPost, "remark42": remark42}
	if err := s.Templates.ExecuteTemplate(w, "post", data); err != nil {
		return err
	}

	return nil
}
