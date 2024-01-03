package client

import (
	"fmt"
	"io"
	"net/http"

	"github.com/quantonganh/blog"
	"github.com/rs/zerolog/hlog"
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

func (n *Newsletter) Subscribe(r *http.Request, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/subscriptions", n.baseURL), body)
	if err != nil {
		return nil, err
	}

	requestID, ok := hlog.IDFromRequest(r)
	if ok {
		req.Header.Set("Request-Id", requestID.String())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}

func (n *Newsletter) Confirm(r *http.Request, token string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/subscriptions/confirm?token=%s", n.baseURL, token), nil)
	if err != nil {
		return nil, err
	}

	requestID, ok := hlog.IDFromRequest(r)
	if ok {
		req.Header.Set("Request-Id", requestID.String())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}

func (n *Newsletter) Unsubscribe(r *http.Request, email, hash string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/unsubscribe?email=%s&hash=%s", n.baseURL, email, hash), nil)
	if err != nil {
		return nil, err
	}

	requestID, ok := hlog.IDFromRequest(r)
	if ok {
		req.Header.Set("Request-Id", requestID.String())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}
