package main

import (
	"log"
	"net/http"

	"github.com/quantonganh/blog/subscriber/hash"
)

const unsubscribeMessage = "Unsubscribed"

func (a *app) unsubscribeHandler(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	email := query.Get("email")
	hashValue := query.Get("hash")
	expectedHash, err := hash.ComputeHmac256(email, a.Config.Newsletter.HMAC.Secret)
	if err != nil {
		log.Fatal(err)
	}

	if hashValue == expectedHash {
		if err := a.MailingList.Unsubscribe(email); err != nil {
			return err
		}

		return a.renderSubscribe(w, unsubscribeMessage)
	}

	return a.renderSubscribe(w, "Either email or hash is invalid.")
}
