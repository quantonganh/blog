package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/feeds"
	"github.com/pkg/errors"
)

func (s *Server) rssHandler(w http.ResponseWriter, r *http.Request) error {
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
				Href: fmt.Sprintf("%s%s", s.URL(), p.URI),
			},
			Description: p.Description,
			Created:     p.Date.Time,
		})
	}
	feed.Items = items

	rss, err := feed.ToRss()
	if err != nil {
		return errors.Wrapf(err, "failed to create RSS")
	}

	_, err = w.Write([]byte(rss))
	if err != nil {
		return err
	}

	return nil
}
