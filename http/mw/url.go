package mw

import (
	"net/http"
)

// URLHost sets URL host
// https://github.com/gorilla/handlers/issues/177
func URLHost(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		r.URL.Host = r.Host
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
