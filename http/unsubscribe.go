package http

import (
	"net/http"

	"github.com/quantonganh/blog/pkg/hash"
)

const (
	unsubscribeMessage        = "Unsubscribed"
	invalidUnsubscribeMessage = "Either email or hash is invalid."
)

func (s *Server) unsubscribeHandler(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	email := query.Get("email")
	hashValue := query.Get("hash")
	expectedHash, err := hash.ComputeHmac256(email, s.SMTPService.GetHMACSecret())
	if err != nil {
		return err
	}

	if hashValue == expectedHash {
		if err := s.SubscribeService.Unsubscribe(email); err != nil {
			return err
		}

		if err := s.Renderer.RenderResponseMessage(w, unsubscribeMessage); err != nil {
			return err
		}
	} else {
		if err := s.Renderer.RenderResponseMessage(w, invalidUnsubscribeMessage); err != nil {
			return err
		}
	}

	return nil
}
