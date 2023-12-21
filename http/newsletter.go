package http

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/quantonganh/mailbus"
)

func (s *Server) subscribeHandler(w http.ResponseWriter, r *http.Request) error {
	email := r.FormValue("email")
	subsReq := mailbus.SubscriptionRequest{
		Email: email,
	}
	req, err := json.Marshal(subsReq)
	if err != nil {
		return err
	}

	resp, err := s.NewsletterService.Subscribe(bytes.NewBuffer(req))
	if err != nil {
		return err
	}

	var subsResp *mailbus.SubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&subsResp); err != nil {
		return err
	}

	if err := s.Renderer.RenderResponseMessage(w, subsResp.Message); err != nil {
		return err
	}

	return nil
}

func (s *Server) confirmHandler(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.Query().Get("token")
	if len(token) == 0 {
		return errors.New("token is not present")
	}

	resp, err := s.NewsletterService.Confirm(token)
	if err != nil {
		return err
	}

	var subsResp *mailbus.SubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&subsResp); err != nil {
		return err
	}

	if err := s.Renderer.RenderResponseMessage(w, subsResp.Message); err != nil {
		return err
	}

	return nil
}

func (s *Server) unsubscribeHandler(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	email := query.Get("email")
	hashValue := query.Get("hash")

	resp, err := s.NewsletterService.Unsubscribe(email, hashValue)
	if err != nil {
		return err
	}

	var subsResp *mailbus.SubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&subsResp); err != nil {
		return err
	}

	if err := s.Renderer.RenderResponseMessage(w, subsResp.Message); err != nil {
		return err
	}

	return nil
}
