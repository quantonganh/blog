package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
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
	hpEmail := r.FormValue("email")
	if hpEmail != "" {
		return NewError(errors.New("better luck next time, bot!"), http.StatusBadRequest, "Congratulations! You've stumbled into our Honeypot field")
	}

	email := r.FormValue("email82244417f9")
	log := zerolog.Ctx(r.Context())
	log.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("email", email)
	})
	a, err := mail.ParseAddress(email)
	if err != nil {
		return NewError(err, http.StatusBadRequest, "Invalid email address.")
	}

	subsReq := map[string]string{
		"url":   s.URL(),
		"email": a.Address,
	}
	body, err := json.Marshal(subsReq)
	if err != nil {
		return err
	}

	resp, err := s.NewsletterService.Subscribe(r, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if err := s.Renderer.RenderResponseMessage(w, contextualClassSuccess, fmt.Sprintf(confirmationMessage, a.Address)); err != nil {
			return err
		}
	case http.StatusUnauthorized:
		if err := s.Renderer.RenderResponseMessage(w, contextualClassWarning, pendingMessage); err != nil {
			return err
		}
	case http.StatusNotFound:
		if err := s.Renderer.RenderResponseMessage(w, contextualClassWarning, fmt.Sprintf(notFoundMessage, a.Address)); err != nil {
			return err
		}
	case http.StatusConflict:
		if err := s.Renderer.RenderResponseMessage(w, contextualClassWarning, alreadySubscribedMessage); err != nil {
			return err
		}
	case http.StatusInternalServerError:
		if err := s.Renderer.RenderResponseMessage(w, contextualClassDanger, errOops); err != nil {
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

	resp, err := s.NewsletterService.Confirm(r, token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == http.StatusOK {
		if err := s.Renderer.RenderResponseMessage(w, contextualClassSuccess, thankyouMessage); err != nil {
			return err
		}
		return nil
	} else if 400 <= statusCode && statusCode <= 499 {
		var m map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
			return err
		}
		return &Error{
			Status:  statusCode,
			Message: m["message"],
		}
	} else {
		var m map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
			return err
		}
		return errors.New(m["error"])
	}
}

func (s *Server) unsubscribeHandler(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	email := query.Get("email")
	hashValue := query.Get("hash")

	resp, err := s.NewsletterService.Unsubscribe(r, email, hashValue)
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
