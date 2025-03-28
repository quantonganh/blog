package http

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/client"
	"github.com/quantonganh/blog/markdown"
	"github.com/quantonganh/blog/ui"
	"github.com/quantonganh/httperror"
)

const (
	shutdownTimeout = 1 * time.Second
)

// Server represents HTTP server
type Server struct {
	logger zerolog.Logger
	ln     net.Listener
	server *http.Server
	router *mux.Router

	Addr   string
	Domain string

	PostService       blog.PostService
	SearchService     blog.SearchService
	Renderer          blog.Renderer
	NewsletterService blog.NewsletterService
	QueueService      blog.QueueService
	EventService      blog.EventService
	StatService       blog.StatService
}

// NewServer create new HTTP server
func NewServer(logger zerolog.Logger, config *blog.Config, posts []*blog.Post) (*Server, error) {
	postService := markdown.NewPostService(posts)
	indexPath := path.Join(path.Dir(config.Posts.Dir), path.Base(config.Posts.Dir)+".bleve")
	searchService, err := markdown.NewSearchService(indexPath, posts)
	if err != nil {
		return nil, err
	}

	s := &Server{
		logger:            logger,
		server:            &http.Server{},
		router:            mux.NewRouter().StrictSlash(true),
		PostService:       postService,
		SearchService:     searchService,
		Renderer:          NewRender(config, postService),
		NewsletterService: client.NewNewsletter(config.Newsletter.BaseURL),
	}

	s.router.Use(hlog.NewHandler(logger))
	s.router.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		if !strings.HasPrefix(r.URL.Path, "/static") && !hasSuffix(r.URL.Path, []string{"ico", "jpg", "jpeg", "png", "gif"}) {
			var event *zerolog.Event
			if 400 <= status && status <= 599 {
				event = hlog.FromRequest(r).Error()
			} else {
				event = hlog.FromRequest(r).Info()
			}
			event.
				Str("method", r.Method).
				Stringer("url", r.URL).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("")

			if config.Env != "local" {
				go func() {
					pattern := `^/\d{4}/\d{2}/\d{2}/[a-z-]+(\.md)?$`
					regex := regexp.MustCompile(pattern)
					ua := r.Header.Get("User-Agent")
					if regex.MatchString(r.URL.Path) && !strings.Contains(strings.ToLower(ua), "bot") {
						ip, err := httperror.GetIP(r)
						if err != nil {
							s.logger.Error().Err(err).Msg("failed to get IP address")
							return
						}
						userID := generateUserID(ip, ua)

						urlPath := r.URL.Path
						if !strings.HasSuffix(urlPath, ".md") {
							urlPath += ".md"
						}
						referer := r.Header.Get("Referer")
						if referer == "" {
							referer = "Unknown"
						}
						now := time.Now().Format("2006-01-02T15:04:05Z")
						data := map[string]string{
							"ip":         ip,
							"user_agent": ua,
							"url":        urlPath,
							"referer":    referer,
							"time":       now,
						}
						jsonData, err := json.Marshal(data)
						if err != nil {
							s.logger.Error().Err(err).Msg("failed to encode message value")
							return
						}
						if err := s.EventService.SendMessage("page-views", userID, jsonData); err != nil {
							s.logger.Error().Err(err).Msg("error sending message")
							return
						}
					}
				}()
			}
		}
	}))
	s.router.Use(httperror.RealIPHandler("ip"))
	s.router.Use(hlog.UserAgentHandler("user_agent"))
	s.router.Use(hlog.RefererHandler("referer"))
	s.router.Use(hlog.RequestIDHandler("req_id", "Request-Id"))

	sentryHandler := sentryhttp.New(sentryhttp.Options{})
	s.router.Use(sentryHandler.Handle)

	s.server.Handler = http.HandlerFunc(s.serveHTTP)

	s.newRoute("/favicon.ico", faviconHandler)
	s.newRoute("/", s.homeHandler)
	s.router.NotFoundHandler = s.Error(s.homeHandler)
	s.newRoute("/{year:20[0-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}/{postName}", s.postHandler(config.Posts.Dir))
	s.newRoute("/{year:20[0-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}", s.postsByDateHandler)
	s.newRoute("/{year:20[0-9][0-9]}/{month:0[1-9]|1[012]}", s.postsByMonthHandler)
	s.newRoute("/{year:20[0-9][0-9]}", s.postsByYearHandler)
	s.router.PathPrefix("/about").HandlerFunc(s.Error(s.postHandler(config.Posts.Dir)))
	s.router.PathPrefix("/resume").HandlerFunc(s.Error(s.postHandler(config.Posts.Dir)))
	s.router.PathPrefix("/projects").HandlerFunc(s.Error(s.postHandler(config.Posts.Dir)))
	s.router.PathPrefix("/uses").HandlerFunc(s.Error(s.postHandler(config.Posts.Dir)))
	s.router.PathPrefix("/now").HandlerFunc(s.Error(s.postHandler(config.Posts.Dir)))
	s.newRoute("/photos", s.photosHandler)
	s.newRoute("/categories/{categoryName}", s.categoryHandler)
	s.newRoute("/tags", s.tagsHandler)
	s.newRoute("/archives", s.archivesHandler)
	s.newRoute("/tags/{tagName}", s.tagHandler)
	s.router.PathPrefix("/static/").Handler(http.FileServer(http.FS(ui.StaticFS)))
	s.newRoute("/search", s.searchHandler)
	s.newRoute("/sitemap.xml", s.sitemapHandler)
	s.newRoute("/rss.xml", s.rssHandler)

	s.newRoute("/subscriptions", s.subscribeHandler).Methods(http.MethodPost)
	subRouter := s.router.PathPrefix("/subscriptions").Subrouter()
	subRouter.HandleFunc("/confirm", s.Error(s.confirmHandler))
	s.newRoute("/unsubscribe", s.unsubscribeHandler)

	if config.Env != "local" {
		s.newRoute("/webhook", s.webhookHandler(config)).Methods(http.MethodPost)
		s.newRoute("/stats", s.statsHandler)
	}

	return s, nil
}

func (s *Server) newRoute(path string, h appHandler) *mux.Route {
	return s.router.HandleFunc(path, s.Error(h))
}

func generateUserID(ip, ua string) string {
	encodedStr := base64.StdEncoding.EncodeToString([]byte(ip + ua))

	hash := md5.Sum([]byte(encodedStr))

	return fmt.Sprintf("%x", hash)
}

// Scheme returns scheme
func (s *Server) Scheme() string {
	if s.UseTLS() {
		return "https"
	}
	return "http"
}

// UseTLS checks if server use TLS or not
func (s *Server) UseTLS() bool {
	return s.Domain != "localhost"
}

// Port returns server port
func (s *Server) Port() int {
	if s.ln == nil {
		return 0
	}
	return s.ln.Addr().(*net.TCPAddr).Port
}

// URL returns server URL
func (s *Server) URL() string {
	domain := "localhost"
	if s.Domain != "" {
		domain = s.Domain
	}

	if flag.Lookup("test.v") != nil {
		return fmt.Sprintf("http://%s", domain)
	}

	scheme := s.Scheme()
	if s.Scheme() == "https" {
		return fmt.Sprintf("%s://%s", scheme, domain)
	}
	return fmt.Sprintf("%s://%s:%d", scheme, domain, s.Port())
}

func faviconHandler(w http.ResponseWriter, r *http.Request) error {
	file, _ := ui.StaticFS.ReadFile("static/favicon.ico")
	_, err := w.Write(file)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) error {
	if err := s.Renderer.RenderPosts(w, r, s.PostService.GetAllPosts()); err != nil {
		return err
	}

	return nil
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Open opens a connection to HTTP server
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

// Close shutdowns HTTP server
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}
