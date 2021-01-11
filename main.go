package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/astaxie/beego/utils/pagination"
	"github.com/blevesearch/bleve"
	"github.com/flosch/pongo2"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/quantonganh/blog/subscriber/mail"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/quantonganh/blog/config"
	"github.com/quantonganh/blog/post"
	"github.com/quantonganh/blog/subscriber"
)

const (
	defaultPostsPerPage = 10
	xmlns               = "http://www.sitemaps.org/schemas/sitemap/0.9"
	indexPath           = "posts.bleve"
	connectionTimeout   = 10 * time.Second
)

var (
	funcMap = template.FuncMap{
		"toISODate": post.ToISODate,
	}
	templates = template.Must(
		template.New("").Funcs(funcMap).ParseGlob("templates/*.html"),
	)
)

type app struct {
	*config.Config
	post.Blog
	subscriber.MailingList
	mail.Mailer
}

func main() {
	posts, err := post.GetAllPosts("posts")
	if err != nil {
		log.Fatal(err)
	}

	blog := post.NewBlog(posts)
	a := app{
		Blog: blog,
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	var cfg *config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}
	a.Config = cfg

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = client.Disconnect(ctx)
	}()

	db := client.Database("mailing_list")
	ml := subscriber.NewMailingList(db)
	a.MailingList = ml

	pageURL, err := url.Parse(os.Getenv("PAGE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	a.Mailer = mail.NewGmail(pageURL, cfg, templates, ml)
	latestPosts := blog.GetLatestPosts(a.Config.Newsletter.Frequency)
	a.Mailer.SendNewsletter(latestPosts)

	router := mux.NewRouter()
	router.HandleFunc("/favicon.ico", faviconHandler)
	router.HandleFunc("/", mwError(a.homeHandler(posts)))
	router.NotFoundHandler = mwError(a.homeHandler(posts))
	router.HandleFunc("/{year:20[1-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}/{postName}", mwError(a.postHandler))
	router.HandleFunc("/tag/{tagName}", mwError(a.tagHandler))
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets", http.FileServer(http.Dir("assets"))))
	router.HandleFunc("/search", mwError(a.searchHandler))
	router.HandleFunc("/sitemap.xml", mwError(a.sitemapHandler(posts)))
	router.HandleFunc("/rss.xml", mwError(a.rssHandler(posts)))

	router.HandleFunc("/subscribe", mwError(a.subscribeHandler(uuid.NewV4().String()))).Methods(http.MethodPost)
	s := router.PathPrefix("/subscribe").Subrouter()
	s.HandleFunc("/confirm", mwError(a.confirmHandler))
	router.HandleFunc("/unsubscribe", mwError(a.unsubscribeHandler))

	loggingHandler := handlers.ProxyHeaders(handlers.LoggingHandler(os.Stdout, router))
	log.Fatal(http.ListenAndServe(":80", mwURLHost(loggingHandler)))
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

// https://github.com/gorilla/handlers/issues/177
func mwURLHost(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		r.URL.Host = r.Host
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

func mwError(hf handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := hf(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (a *app) homeHandler(posts []*post.Post) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		return a.renderHTML(w, r, posts)
	}
}

func (a *app) postHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	year := vars["year"]
	month := vars["month"]
	fileName := vars["postName"]

	currentPost, err := post.ParseMarkdown(context.Background(), filepath.Join("posts", year, month, fileName+".md"))
	if err != nil {
		return err
	}

	relatedPosts, err := a.Blog.GetRelatedPosts(currentPost)
	if err != nil {
		return err
	}

	previousPost, nextPost := a.Blog.GetPreviousAndNextPost(currentPost)
	if previousPost != nil {
		currentPost.HasPrev = true
	}
	if nextPost != nil {
		currentPost.HasNext = true
	}

	remark42, err := getRemarkURL()
	if err != nil {
		return err
	}

	data := pongo2.Context{"title": currentPost.Title, "currentPost": currentPost, "relatedPosts": relatedPosts, "previousPost": previousPost, "nextPost": nextPost, "remark42": remark42}
	if a.Config != nil {
		data["navbarItems"] = a.Config.Navbar.Items
	}
	if err := templates.ExecuteTemplate(w, "post", data); err != nil {
		return err
	}

	return nil
}

func (a *app) tagHandler(w http.ResponseWriter, r *http.Request) error {
	tag := mux.Vars(r)["tagName"]

	postsByTag, err := a.Blog.GetPostsByTag(tag)
	if err != nil {
		return err
	}

	if err := a.renderHTML(w, r, postsByTag); err != nil {
		return err
	}

	return nil
}

func (a *app) searchHandler(w http.ResponseWriter, r *http.Request) error {
	var (
		index bleve.Index
		err   error
	)
	if _, err = os.Stat(indexPath); os.IsNotExist(err) {
		index, err = a.Blog.IndexPosts(indexPath)
		if err != nil {
			return errors.Errorf("failed to index posts: %v", err)
		}
	} else if err == nil {
		index, err = bleve.OpenUsing(indexPath, map[string]interface{}{
			"read_only": true,
		})
		if err != nil {
			return errors.Errorf("failed to open index at %s: %v", indexPath, err)
		}
	}
	defer func() {
		_ = index.Close()
	}()

	if err := r.ParseForm(); err != nil {
		return errors.Errorf("failed to parse form: %v", err)
	}

	searchPosts, err := a.Blog.Search(index, r.FormValue("q"))
	if err != nil {
		return errors.Errorf("failed to search: %v", err)
	}

	if err := a.renderHTML(w, r, searchPosts); err != nil {
		return errors.Errorf("failed to render HTML: %v", err)
	}

	return nil
}

func (a *app) renderHTML(w http.ResponseWriter, r *http.Request, posts []*post.Post) error {
	var (
		postsPerPage int
		err          error
	)
	postsPerPageEnv, exists := os.LookupEnv("POSTS_PER_PAGE")
	if !exists {
		postsPerPage = defaultPostsPerPage
	} else {
		postsPerPage, err = strconv.Atoi(postsPerPageEnv)
		if err != nil {
			return errors.Errorf("failed to convert %s to int: %v", postsPerPageEnv, err)
		}
	}

	nums := len(posts)
	paginator := pagination.NewPaginator(r, postsPerPage, int64(nums))
	offset := paginator.Offset()

	endPos := offset + postsPerPage
	if endPos > nums {
		endPos = nums
	}

	data := pongo2.Context{"posts": posts[offset:endPos], "paginator": paginator}
	if a.Config != nil {
		data["navbarItems"] = a.Config.Navbar.Items
	}
	if err := templates.ExecuteTemplate(w, "home", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}

type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

func (a *app) sitemapHandler(posts []*post.Post) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		scheme := "http"
		if xForwardedProto := r.Header.Get("X-Forwarded-Proto"); xForwardedProto != "" {
			scheme = xForwardedProto
		}

		sitemap := Sitemap{
			XMLNS: xmlns,
			URLs: []URL{
				{
					Loc: fmt.Sprintf("%s://%s", scheme, r.Host),
				},
			},
		}

		for _, p := range posts {
			sitemap.URLs = append(sitemap.URLs, URL{
				Loc:     fmt.Sprintf("%s://%s/%s", scheme, r.Host, p.URI),
				LastMod: post.ToISODate(p.Date),
			})
		}

		output, err := xml.MarshalIndent(sitemap, "  ", "    ")
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(xml.Header + string(output)))
		if err != nil {
			return err
		}

		return nil
	}
}
