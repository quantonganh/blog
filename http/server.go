package http

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/http/html"
	"github.com/quantonganh/blog/ondisk"
)

const (
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

//go:embed html/templates/*.html
var htmlFiles embed.FS

//go:embed assets
var assets embed.FS

func NewServer(config *blog.Config, posts []*blog.Post, indexPath string) (*Server, error) {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: config.Sentry.DSN,
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to init Sentry, DSN: %s", config.Sentry.DSN)
	}
	defer sentry.Flush(2 * time.Second)

	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	funcMap := template.FuncMap{
		"toISODate":   blog.ToISODate,
		"toMonthName": blog.ToMonthName,
	}
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(htmlFiles, "html/templates/*.html"))
	postService, err := ondisk.NewPostService(posts, indexPath)
	if err != nil {
		return nil, err
	}
	s := &Server{
		server:      &http.Server{},
		router:      mux.NewRouter().StrictSlash(true),
		PostService: postService,
		Renderer:    html.NewRender(config, postService, tmpl),
	}

	zlog := zerolog.New(os.Stdout).With().
		Timestamp().
		Logger()
	s.router.Use(hlog.NewHandler(zlog))
	s.router.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	s.router.Use(hlog.RemoteAddrHandler("ip"))
	s.router.Use(hlog.UserAgentHandler("user_agent"))
	s.router.Use(hlog.RefererHandler("referer"))
	s.router.Use(hlog.RequestIDHandler("req_id", "Request-Id"))

	s.router.Use(sentryHandler.Handle)

	s.server.Handler = http.HandlerFunc(s.serveHTTP)

	s.router.HandleFunc("/favicon.ico", s.Error(faviconHandler))
	s.router.HandleFunc("/", s.Error(s.homeHandler))
	s.router.NotFoundHandler = s.Error(s.homeHandler)
	s.router.HandleFunc("/{year:20[0-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}/{postName}", s.Error(s.postHandler(config.Posts.Dir)))
	s.router.HandleFunc("/{year:20[0-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}", s.Error(s.postsByDateHandler))
	s.router.HandleFunc("/{year:20[0-9][0-9]}/{month:0[1-9]|1[012]}", s.Error(s.postsByMonthHandler))
	s.router.HandleFunc("/{year:20[0-9][0-9]}", s.Error(s.postsByYearHandler))
	s.router.HandleFunc("/about", s.Error(s.postHandler(config.Posts.Dir)))
	s.router.HandleFunc("/resume", s.Error(s.postHandler(config.Posts.Dir)))
	s.router.HandleFunc("/photos", s.Error(s.photosHandler))
	s.router.HandleFunc("/category/{categoryName}", s.Error(s.categoryHandler))
	s.router.HandleFunc("/tags", s.Error(s.tagsHandler))
	s.router.HandleFunc("/archives", s.Error(s.archivesHandler))
	s.router.HandleFunc("/tag/{tagName}", s.Error(s.tagHandler))
	s.router.PathPrefix("/assets/").Handler(http.FileServer(http.FS(assets)))
	s.router.HandleFunc("/search", s.Error(s.searchHandler))
	s.router.HandleFunc("/sitemap.xml", s.Error(s.sitemapHandler))
	s.router.HandleFunc("/rss.xml", s.Error(s.rssHandler))

	s.router.HandleFunc("/subscribe", s.Error(s.subscribeHandler)).Methods(http.MethodPost)
	subRouter := s.router.PathPrefix("/subscribe").Subrouter()
	subRouter.HandleFunc("/confirm", s.Error(s.confirmHandler))
	s.router.HandleFunc("/unsubscribe", s.Error(s.unsubscribeHandler))

	return s, nil
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

func faviconHandler(w http.ResponseWriter, r *http.Request) *AppError {
	file, _ := assets.ReadFile("assets/favicon.ico")
	_, err := w.Write(file)
	if err != nil {
		return &AppError{
			Error: err,
			Code:  http.StatusInternalServerError,
		}
	}

	return nil
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
