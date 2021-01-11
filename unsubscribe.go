package main

import (
	"net/http"

	"github.com/pkg/errors"
)

func (b *Blog) unsubscribeHandler(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	email := query.Get("email")
	hash := query.Get("hash")
	expectedHash := computeHmac256(email, b.config.HMAC.Secret)

	if hash == expectedHash {
		subscribers, err := Load(subscribersFile)
		if err != nil {
			return errors.Errorf("failed to load file %subscribers: %v", subscribersFile, err)
		}

		for i, subscriber := range subscribers {
			if subscriber.Email == email {
				subscribers = append(subscribers[:i], subscribers[i+1:]...)
				if err := Save(subscribers, subscribersFile); err != nil {
					return err
				}
				return b.renderSubscribe(w, "Unsubscribed")
			}
		}
	}

	return b.renderSubscribe(w, "Either email or hash is invalid.")
}
