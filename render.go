package blog

import (
	"bytes"
	"net/http"
)

type Renderer interface {
	RenderPosts(w http.ResponseWriter, r *http.Request, posts []*Post) error
	RenderPost(w http.ResponseWriter, currentPost *Post, relatedPosts map[string]*Post, previousPost, nextPost *Post) error
	RenderResponseMessage(w http.ResponseWriter, message string) error
	RenderNewsletter(latestPosts []*Post, serverURL, email string) (*bytes.Buffer, error)
}
