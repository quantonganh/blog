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
	uuid "github.com/satori/go.uuid"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/http/html"
	"github.com/quantonganh/blog/http/mw"
	"github.com/quantonganh/blog/ondisk"
)

const (
	IndexPath       = "posts.bleve"
	shutdownTimeout = 1 * time.Second
)

var (
	funcMap = template.FuncMap{
		"toISODate": blog.ToISODate,
	}
)

type Server struct {
	ln     net.Listener
	server *http.Server
	router *mux.Router

	Addr   string
	Domain string

	Templates *template.Template

	PostService      blog.PostService
	SubscribeService blog.SubscribeService
	SMTPService      blog.SMTPService
}

func NewServer(config *blog.Config, posts []*blog.Post) *Server {
	s := &Server{
		server:      &http.Server{},
		router:      mux.NewRouter(),
		PostService: ondisk.NewPostService(posts),
	}

	s.Templates = template.Must(
		template.New("").Funcs(funcMap).ParseGlob(fmt.Sprintf("%s/*.tmpl", config.Templates.Dir)))

	s.router.Use(logging)

	s.server.Handler = http.HandlerFunc(s.serveHTTP)

	s.router.HandleFunc("/favicon.ico", faviconHandler)
	s.router.HandleFunc("/", mw.Error(s.homeHandler(posts)))
	s.router.NotFoundHandler = mw.Error(s.homeHandler(posts))
	s.router.HandleFunc("/{year:20[1-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}/{postName}", mw.Error(s.postHandler))
	s.router.HandleFunc("/tag/{tagName}", mw.Error(s.tagHandler))
	s.router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets", http.FileServer(http.Dir("http/assets"))))
	s.router.HandleFunc("/search", mw.Error(s.searchHandler(IndexPath)))
	s.router.HandleFunc("/sitemap.xml", mw.Error(s.SitemapHandler(posts)))
	s.router.HandleFunc("/rss.xml", mw.Error(s.rssHandler(posts)))

	s.router.HandleFunc("/subscribe", mw.Error(s.subscribeHandler(uuid.NewV4().String()))).Methods(http.MethodPost)
	subRouter := s.router.PathPrefix("/subscribe").Subrouter()
	subRouter.HandleFunc("/confirm", mw.Error(s.confirmHandler))
	s.router.HandleFunc("/unsubscribe", mw.Error(s.unsubscribeHandler(config.Newsletter.HMAC.Secret)))

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

func (s *Server) homeHandler(posts []*blog.Post) mw.ErrHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		return html.NewPost(s.Templates).Render(w, r, posts)
	}
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
