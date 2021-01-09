package main

import (
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

func (b *Blog) unsubscribeHandler(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	email := query.Get("email")
	hash := query.Get("hash")
	expectedHash := computeHmac256(email, b.config.HMAC.Secret)

	f, err := os.OpenFile(subscribersFile, os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = f.Close()
	}()

	ml := &mailingList{}
	if hash == expectedHash {
		if err := ml.Load(f); err != nil {
			return errors.Errorf("failed to load file %subscribers: %v", subscribersFile, err)
		}

		for i, s := range ml.Subscribers {
			if s.Email == email {
				ml.Subscribers = append(ml.Subscribers[:i], ml.Subscribers[i+1:]...)
				if err := ml.Save(f); err != nil {
					return err
				}
				return b.renderSubscribe(w, "Unsubscribed")
			}
		}
	}

	return b.renderSubscribe(w, "Either email or hash is invalid.")
}
