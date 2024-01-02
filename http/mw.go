package http

import (
	"net/http"
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
			_ = s.Renderer.RenderResponseMessage(w, contextualClassWarning, "Too many requests.")
			return
		}
	})
}
