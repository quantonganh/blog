package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

const (
	contextualClassSuccess = "success"
	contextualClassWarning = "warning"
	contextualClassDanger  = "danger"

	confirmationMessage       = "A confirmation email has been sent to %s. Click the link in the email to confirm and activate your subscription. Check your spam folder if you don't see it within a couple of minutes."
	thankyouMessage           = "Thank you for subscribing to this blog."
	pendingMessage            = "Your subscription status is pending. Please click the confirmation link in your email."
	alreadySubscribedMessage  = "You had been subscribed to this blog already."
	notFoundMessage           = "Cannot found email: %s"
	unsubscribeMessage        = "You've been successfully unsubscribed from our blog updates. If this was unintentional or you change your mind, feel free to resubscribe anytime."
	invalidUnsubscribeMessage = "Either email or hash is invalid"
)

func (s *Server) subscribeHandler(w http.ResponseWriter, r *http.Request) error {
	email := r.FormValue("email")
	subsReq := map[string]string{
		"url":   s.URL(),
		"email": email,
	}
	req, err := json.Marshal(subsReq)
	if err != nil {
		return err
	}

	resp, err := s.NewsletterService.Subscribe(bytes.NewBuffer(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if err := s.Renderer.RenderResponseMessage(w, contextualClassSuccess, fmt.Sprintf(confirmationMessage, email)); err != nil {
			return err
		}
	case http.StatusUnauthorized:
		if err := s.Renderer.RenderResponseMessage(w, contextualClassWarning, pendingMessage); err != nil {
			return err
		}
	case http.StatusNotFound:
		if err := s.Renderer.RenderResponseMessage(w, contextualClassWarning, fmt.Sprintf(notFoundMessage, email)); err != nil {
			return err
		}
	case http.StatusConflict:
		if err := s.Renderer.RenderResponseMessage(w, contextualClassWarning, alreadySubscribedMessage); err != nil {
			return err
		}
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
	defer resp.Body.Close()

	if err := s.Renderer.RenderResponseMessage(w, contextualClassSuccess, thankyouMessage); err != nil {
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
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if err := s.Renderer.RenderResponseMessage(w, contextualClassSuccess, unsubscribeMessage); err != nil {
			return err
		}
	case http.StatusBadRequest:
		if err := s.Renderer.RenderResponseMessage(w, contextualClassWarning, invalidUnsubscribeMessage); err != nil {
			return err
		}
	}

	return nil
}
