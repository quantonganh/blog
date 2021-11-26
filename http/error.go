package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/getsentry/sentry-go"
)

const errOops = "Oops! Something went wrong."

type appHandler func(w http.ResponseWriter, r *http.Request) error

func (s *Server) Error(fn appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err == nil {
			return
		}

		log.Printf("%+v", err)
		sentry.CaptureException(err)

		clientError, ok := err.(ClientError)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_ = s.Renderer.RenderResponseMessage(w, errOops)
			return
		}

		body, err := clientError.Body()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = s.Renderer.RenderResponseMessage(w, errOops)
			return
		}

		status, headers := clientError.Headers()
		for k, v := range headers {
			w.Header().Set(k, v)
		}

		w.WriteHeader(status)

		_, err = w.Write(body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = s.Renderer.RenderResponseMessage(w, errOops)
			return
		}
	}
}

type ClientError interface {
	Error() string
	Body() ([]byte, error)
	Headers() (int, map[string]string)
}

type Error struct {
	Cause   error  `json:"-"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return e.Message + ": " + e.Cause.Error()
}

func (e *Error) Body() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing response body: %v", err)
	}
	return body, nil
}

func (e *Error) Headers() (int, map[string]string) {
	return e.Status, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
}

func NewError(err error, status int, message string) error {
	return &Error{
		Cause:   err,
		Message: message,
		Status:  status,
	}
}
