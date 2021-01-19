package http

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
)

type AppError struct {
	Error   error
	Message string
	Code    int
}

type ErrHandlerFunc func(w http.ResponseWriter, r *http.Request) *AppError

func (s *Server) Error(hf ErrHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if e := hf(w, r); e != nil {
			if e.Message == "" {
				e.Message = "An error has occurred."
			}
			fmt.Printf("%+v\n", e.Error)
			sentry.CaptureException(e.Error)
			w.WriteHeader(e.Code)
			_ = s.Renderer.RenderResponseMessage(w, e.Message)
		}
	}
}
