package blog

import (
	"io"
	"net/http"
)

type NewsletterService interface {
	Subscribe(body io.Reader) (*http.Response, error)
	Confirm(token string) (*http.Response, error)
	Unsubscribe(email, hash string) (*http.Response, error)
}
