package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/feeds"
	"github.com/pkg/errors"
	"github.com/quantonganh/blog/post"
)

func (a *app) rssHandler(posts []*post.Post) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		scheme := "http"
		if xForwardedProto := r.Header.Get("X-Forwarded-Proto"); xForwardedProto != "" {
			scheme = xForwardedProto
		}

		feed := &feeds.Feed{
			Title: fmt.Sprintf("%s app", r.Host),
			Link: &feeds.Link{
				Href: fmt.Sprintf("%s://%s", scheme, r.Host),
			},
		}

		var items []*feeds.Item
		for _, p := range posts {
			items = append(items, &feeds.Item{
				Title: p.Title,
				Link: &feeds.Link{
					Href: fmt.Sprintf("%s://%s/%s", scheme, r.Host, p.URI),
				},
				Description: p.Description,
				Created:     p.Date.Time,
			})
		}
		feed.Items = items

		rss, err := feed.ToRss()
		if err != nil {
			return errors.Errorf("failed to create RSS: %v", err)
		}

		_, err = w.Write([]byte(rss))
		if err != nil {
			return err
		}

		return nil
	}
}
