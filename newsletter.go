package blog

import (
	"io"
	"net/http"
)

type NewsletterService interface {
	Subscribe(r *http.Request, body io.Reader) (*http.Response, error)
	Confirm(r *http.Request, token string) (*http.Response, error)
	Unsubscribe(r *http.Request, email, hash string) (*http.Response, error)
	Send(r *http.Request, body io.Reader) (*http.Response, error)
}
