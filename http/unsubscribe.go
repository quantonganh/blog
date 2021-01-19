package http

import (
	"log"
	"net/http"

	"github.com/quantonganh/blog/pkg/hash"
)

const (
	unsubscribeMessage        = "Unsubscribed"
	invalidUnsubscribeMessage = "Either email or hash is invalid."
)

func (s *Server) unsubscribeHandler(w http.ResponseWriter, r *http.Request) *AppError {
	query := r.URL.Query()
	email := query.Get("email")
	hashValue := query.Get("hash")
	expectedHash, err := hash.ComputeHmac256(email, s.SMTPService.GetHMACSecret())
	if err != nil {
		log.Fatal(err)
	}

	if hashValue == expectedHash {
		if err := s.SubscribeService.Unsubscribe(email); err != nil {
			return &AppError{
				Error: err,
				Code:  http.StatusInternalServerError,
			}
		}

		if err := s.Renderer.RenderResponseMessage(w, unsubscribeMessage); err != nil {
			return &AppError{
				Error:   err,
				Message: "Failed to render subscribe template.",
				Code:    http.StatusInternalServerError,
			}
		}
	}

	if err := s.Renderer.RenderResponseMessage(w, invalidUnsubscribeMessage); err != nil {
		return &AppError{
			Error:   err,
			Message: "Failed to render subscribe template.",
			Code:    http.StatusInternalServerError,
		}
	}

	return nil
}
