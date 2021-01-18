package http

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/http/html"
	"github.com/quantonganh/blog/http/mw"
	"github.com/quantonganh/blog/ondisk"
)

const (
	IndexPath       = "posts.bleve"
	shutdownTimeout = 1 * time.Second
)

type Server struct {
	ln     net.Listener
	server *http.Server
	router *mux.Router

	Addr   string
	Domain string

	PostService      blog.PostService
	SubscribeService blog.SubscribeService
	SMTPService      blog.SMTPService
	Renderer         blog.Renderer
}

func NewServer(config *blog.Config, posts []*blog.Post) *Server {
	funcMap := template.FuncMap{
		"toISODate": blog.ToISODate,
	}
	tmpl := template.Must(
		template.New("").Funcs(funcMap).ParseGlob(fmt.Sprintf("%s/*.html", config.Templates.Dir)))
	s := &Server{
		server:      &http.Server{},
		router:      mux.NewRouter(),
		PostService: ondisk.NewPostService(posts),
		Renderer:    html.NewRender(config, tmpl),
	}

	s.router.Use(logging)

	s.server.Handler = http.HandlerFunc(s.serveHTTP)

	s.router.HandleFunc("/favicon.ico", faviconHandler)
	s.router.HandleFunc("/", s.Error(s.homeHandler))
	s.router.NotFoundHandler = s.Error(s.homeHandler)
	s.router.HandleFunc("/{year:20[1-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}/{postName}", s.Error(s.postHandler))
	s.router.HandleFunc("/tag/{tagName}", s.Error(s.tagHandler))
	s.router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets", http.FileServer(http.Dir("http/assets"))))
	s.router.HandleFunc("/search", s.Error(s.searchHandler(IndexPath)))
	s.router.HandleFunc("/sitemap.xml", s.Error(s.sitemapHandler))
	s.router.HandleFunc("/rss.xml", s.Error(s.rssHandler))

	s.router.HandleFunc("/subscribe", s.Error(s.subscribeHandler)).Methods(http.MethodPost)
	subRouter := s.router.PathPrefix("/subscribe").Subrouter()
	subRouter.HandleFunc("/confirm", s.Error(s.confirmHandler))
	s.router.HandleFunc("/unsubscribe", s.Error(s.unsubscribeHandler))

	return s
}

func (s *Server) Scheme() string {
	if s.UseTLS() {
		return "https"
	}
	return "http"
}

func (s *Server) UseTLS() bool {
	return s.Domain != ""
}

func (s *Server) Port() int {
	if s.ln == nil {
		return 0
	}
	return s.ln.Addr().(*net.TCPAddr).Port
}

func (s *Server) URL() string {
	scheme, port := s.Scheme(), s.Port()

	domain := "localhost"
	if s.Domain != "" {
		domain = s.Domain
	}

	if port == 80 || port == 443 {
		return fmt.Sprintf("%s://%s", scheme, domain)
	}
	return fmt.Sprintf("%s://%s:%d", scheme, domain, s.Port())
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) *AppError {
	if err := s.Renderer.RenderPosts(w, r, s.PostService.GetAllPosts()); err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	return nil
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) Open() (err error) {
	s.ln, err = net.Listen("tcp", s.Addr)
	if err != nil {
		return errors.Errorf("failed to listen to port %s: %v", s.Addr, err)
	}

	go func() {
		_ = s.server.Serve(s.ln)
	}()

	return nil
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func logging(next http.Handler) http.Handler {
	return mw.URLHost(handlers.ProxyHeaders(handlers.LoggingHandler(os.Stdout, next)))
}
