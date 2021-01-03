package main

import (
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/astaxie/beego/utils/pagination"
	"github.com/blevesearch/bleve"
	"github.com/flosch/pongo2"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/quantonganh/blog/config"
)

const (
	title               = "Learning makes me happy"
	defaultPostsPerPage = 10
	xmlns               = "http://www.sitemaps.org/schemas/sitemap/0.9"
	indexPath           = "posts.bleve"
)

var (
	funcMap = template.FuncMap{
		"toISODate": toISODate,
	}
	templates = template.Must(
		template.New("").Funcs(funcMap).ParseGlob("templates/*.html"),
	)
)

type Blog struct {
	config *config.Config
	posts  []*Post
}

type Post struct {
	URI         string
	Title       string
	Date        publishDate
	Description string
	Content     template.HTML
	Tags        []string
	HasPrev     bool
	HasNext     bool
}

func main() {
	var (
		configPath string
		cfg        *config.Config
		err        error
	)

	posts, err := getAllPosts("posts/**/*.md")
	if err != nil {
		log.Fatal(err)
	}
	b := Blog{
		posts: posts,
	}

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	flagSet := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagSet[f.Name] = true })

	if flagSet["config"] {
		cfg, err = config.NewConfig(configPath)
		if err != nil {
			log.Fatal(err)
		}
		b.config = cfg
	}

	router := mux.NewRouter()
	router.HandleFunc("/favicon.ico", faviconHandler)
	router.HandleFunc("/", mwError(b.homeHandler))
	router.NotFoundHandler = mwError(b.homeHandler)
	router.HandleFunc("/{year:20[1-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}/{postName}", mwError(b.postHandler))
	router.HandleFunc("/tag/{tagName}", mwError(b.tagHandler))
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets", http.FileServer(http.Dir("assets"))))
	router.HandleFunc("/search", mwError(b.searchHandler))
	router.HandleFunc("/sitemap.xml", mwError(b.sitemapHandler))

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

func (b *Blog) homeHandler(w http.ResponseWriter, r *http.Request) error {
	return b.renderHTML(w, r, b.posts)
}

func (b *Blog) postHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	year := vars["year"]
	month := vars["month"]
	fileName := vars["postName"]

	currentPost, err := parseMarkdown(context.Background(), filepath.Join("posts", year, month, fileName+".md"))
	if err != nil {
		return err
	}

	relatedPosts, err := getRelatedPosts(b.posts, currentPost)
	if err != nil {
		return err
	}

	previousPost, nextPost := getPreviousAndNextPost(b.posts, currentPost)
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
	if b.config != nil {
		data["navbarItems"] = b.config.Navbar.Items
	}
	if err := templates.ExecuteTemplate(w, "post", data); err != nil {
		return err
	}

	return nil
}

func (b *Blog) tagHandler(w http.ResponseWriter, r *http.Request) error {
	tag := mux.Vars(r)["tagName"]

	postsByTag, err := getPostsByTag(b.posts, tag)
	if err != nil {
		return err
	}

	if err := b.renderHTML(w, r, postsByTag); err != nil {
		return err
	}

	return nil
}

func (b *Blog) searchHandler(w http.ResponseWriter, r *http.Request) error {
	var (
		index bleve.Index
		err   error
	)
	if _, err = os.Stat(indexPath); os.IsNotExist(err) {
		index, err = b.indexPosts(indexPath)
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

	searchPosts, err := b.search(index, r.FormValue("q"))
	if err != nil {
		return errors.Errorf("failed to search: %v", err)
	}

	if err := b.renderHTML(w, r, searchPosts); err != nil {
		return errors.Errorf("failed to render HTML: %v", err)
	}

	return nil
}

func (b *Blog) renderHTML(w http.ResponseWriter, r *http.Request, posts []*Post) error {
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

	data := pongo2.Context{"title": title, "posts": posts[offset:endPos], "paginator": paginator}
	if b.config != nil {
		data["navbarItems"] = b.config.Navbar.Items
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

func (b *Blog) sitemapHandler(w http.ResponseWriter, r *http.Request) error {
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

	for _, post := range b.posts {
		sitemap.URLs = append(sitemap.URLs, URL{
			Loc:     fmt.Sprintf("%s://%s/%s", scheme, r.Host, post.URI),
			LastMod: toISODate(post.Date),
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
