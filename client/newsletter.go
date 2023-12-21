package client

import (
	"fmt"
	"io"
	"net/http"

	"github.com/quantonganh/blog"
)

type Newsletter struct {
	baseURL    string
	httpClient *http.Client
}

func NewNewsletter(baseURL string) blog.NewsletterService {
	return &Newsletter{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (n *Newsletter) Subscribe(body io.Reader) (*http.Response, error) {
	return http.Post(fmt.Sprintf("%s/subscriptions", n.baseURL), "application/json", body)
}

func (n *Newsletter) Confirm(token string) (*http.Response, error) {
	return http.Get(fmt.Sprintf("%s/subscriptions/confirm?token=%s", n.baseURL, token))
}

func (n *Newsletter) Unsubscribe(email, hash string) (*http.Response, error) {
	return http.Get(fmt.Sprintf("%s/unsubscribe?email=%s&hash=%s", n.baseURL, email, hash))
}
