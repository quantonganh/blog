package blog

import (
	"bytes"
	"net/http"
)

// Renderer is the interface that wraps methods related to render HTML page
type Renderer interface {
	RenderPhotos(w http.ResponseWriter) error
	RenderTags(w http.ResponseWriter) error
	RenderArchives(w http.ResponseWriter) error
	RenderPosts(w http.ResponseWriter, r *http.Request, posts []*Post) error
	RenderPost(w http.ResponseWriter, currentPost *Post, relatedPosts []*Post, previousPost, nextPost *Post) error
	RenderResponseMessage(w http.ResponseWriter, message string) error
	RenderNewsletter(latestPosts []*Post, serverURL, email string) (*bytes.Buffer, error)
	RenderVTV(w http.ResponseWriter, letters string, total int, rows [][]string) error
}
