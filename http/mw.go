package http

import (
	"net/http"

	"github.com/rs/zerolog"
)

type customResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *customResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *customResponseWriter) StatusCode() int {
	return w.statusCode
}

func (s *Server) toHTML(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customWriter := &customResponseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(customWriter, r)

		if customWriter.StatusCode() == http.StatusTooManyRequests {
			log := zerolog.Ctx(r.Context())
			log.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("email", r.FormValue("email"))
			})
			_ = s.Renderer.RenderResponseMessage(w, contextualClassWarning, "Too many requests.")
			return
		}
	})
}
