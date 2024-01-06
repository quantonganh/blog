package http

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
)

const (
	errOops = "Oops! Something went wrong. Please try again later."
)

type appHandler func(w http.ResponseWriter, r *http.Request) error

// Error parse HTTP error and write to header and body
func (s *Server) Error(fn appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err == nil {
			return
		}

		log := zerolog.Ctx(r.Context())
		log.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str(zerolog.ErrorFieldName, err.Error())
		})

		clientError, ok := err.(ClientError)
		if !ok {
			sentry.CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = s.Renderer.RenderResponseMessage(w, contextualClassDanger, errOops)
			return
		}

		status, _ := clientError.Headers()
		w.WriteHeader(status)
		_ = s.Renderer.RenderResponseMessage(w, contextualClassWarning, clientError.Body())
	}
}

// ClientError is the interface that wraps methods related to error on the client side
type ClientError interface {
	Error() string
	Body() string
	Headers() (int, map[string]string)
}

// Error represents a detail error message
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

// Body returns response body from error
func (e *Error) Body() string {
	return e.Message
}

// Headers returns status and header
func (e *Error) Headers() (int, map[string]string) {
	return e.Status, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
}

// NewError returns new error message
func NewError(err error, status int, message string) error {
	return &Error{
		Cause:   err,
		Message: message,
		Status:  status,
	}
}
