package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/feeds"
)

func (s *Server) rssHandler(w http.ResponseWriter, r *http.Request) *AppError {
	feed := &feeds.Feed{
		Title: fmt.Sprintf("%s app", r.Host),
		Link: &feeds.Link{
			Href: s.URL(),
		},
	}

	var items []*feeds.Item
	for _, p := range s.PostService.GetAllPosts() {
		items = append(items, &feeds.Item{
			Title: p.Title,
			Link: &feeds.Link{
				Href: fmt.Sprintf("%s/%s", s.URL(), p.URI),
			},
			Description: p.Description,
			Created:     p.Date.Time,
		})
	}
	feed.Items = items

	rss, err := feed.ToRss()
	if err != nil {
		return &AppError{
			Error:   err,
			Message: "failed to create RSS",
			Code:    http.StatusInternalServerError,
		}
	}

	_, err = w.Write([]byte(rss))
	if err != nil {
		return &AppError{
			Error:   err,
			Message: "Can't write the data to the connection as part of a HTTP reply",
			Code:    http.StatusInternalServerError,
		}
	}

	return nil
}
