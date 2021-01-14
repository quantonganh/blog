package mw

import (
	"net/http"
)

type ErrHandlerFunc func(w http.ResponseWriter, r *http.Request) error

func Error(hf ErrHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := hf(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
