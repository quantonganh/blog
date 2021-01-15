package http

import (
	"log"
	"net/http"

	"github.com/quantonganh/blog/http/mw"
	"github.com/quantonganh/blog/pkg/hash"
)

const (
	unsubscribeMessage        = "Unsubscribed"
	invalidUnsubscribeMessage = "Either email or hash is invalid."
)

func (s *Server) unsubscribeHandler(hmacSecret string) mw.ErrHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		query := r.URL.Query()
		email := query.Get("email")
		hashValue := query.Get("hash")
		expectedHash, err := hash.ComputeHmac256(email, hmacSecret)
		if err != nil {
			log.Fatal(err)
		}

		if hashValue == expectedHash {
			if err := s.SubscribeService.Unsubscribe(email); err != nil {
				return err
			}

			return s.Renderer.RenderSubscribeMessage(w, unsubscribeMessage)
		}

		return s.Renderer.RenderSubscribeMessage(w, invalidUnsubscribeMessage)
	}
}
