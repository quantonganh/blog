package main

import (
	"context"
	"encoding/xml"
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
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
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

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

func mwError(hf handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := hf(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	posts, err := getAllPosts("posts/**/*.md")
	if err != nil {
		log.Fatal(err)
	}
	b := Blog{
		posts: posts,
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

	log.Fatal(http.ListenAndServe(":80", logHandler(router)))
}

type Blog struct {
	posts []*Post
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

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

func (b *Blog) homeHandler(w http.ResponseWriter, r *http.Request) error {
	return renderHTML(w, r, b.posts)
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

	if err := renderHTML(w, r, postsByTag); err != nil {
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

	searchPosts, err := b.search(index, r.FormValue("search"))
	if err != nil {
		return errors.Errorf("failed to search: %v", err)
	}

	if err := renderHTML(w, r, searchPosts); err != nil {
		return errors.Errorf("failed to render HTML: %v", err)
	}

	return nil
}

func renderHTML(w http.ResponseWriter, r *http.Request, posts []*Post) error {
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

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}

	n, err := w.ResponseWriter.Write(b)
	w.length += n

	return n, err
}

func logHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w}
		handler.ServeHTTP(sw, r)
		log.Printf("%s %s %s %d %d", r.RemoteAddr, r.Method, r.URL.Path, sw.status, sw.length)
	})
}
