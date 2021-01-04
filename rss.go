package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/feeds"
	"github.com/pkg/errors"
)

func (b *Blog) rssHandler(w http.ResponseWriter, r *http.Request) error {
	scheme := "http"
	if xForwardedProto := r.Header.Get("X-Forwarded-Proto"); xForwardedProto != "" {
		scheme = xForwardedProto
	}

	feed := &feeds.Feed{
		Title: fmt.Sprintf("%s blog", r.Host),
		Link: &feeds.Link{
			Href: fmt.Sprintf("%s://%s", scheme, r.Host),
		},
	}

	var items []*feeds.Item
	for _, post := range b.posts {
		items = append(items, &feeds.Item{
			Title: post.Title,
			Link: &feeds.Link{
				Href: fmt.Sprintf("%s://%s/%s", scheme, r.Host, post.URI),
			},
			Description: post.Description,
			Created:     post.Date.Time,
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
