package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Depado/bfchroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/styles"
	"github.com/astaxie/beego/utils/pagination"
	"github.com/bmatcuk/doublestar/v2"
	"github.com/flosch/pongo2"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	bf "gopkg.in/russross/blackfriday.v2"
	"gopkg.in/yaml.v2"
)

const (
	title               = "Learning makes me happy"
	yamlSeparator       = "---"
	layoutUnix          = "Mon Jan 2 15:04:05 -07 2006"
	layoutISO           = "2006-01-02"
	defaultPostsPerPage = 10
	xmlns               = "http://www.sitemaps.org/schemas/sitemap/0.9"
)

var (
	funcMap = template.FuncMap{
		"toISODate": toISODate,
	}
	templates = template.Must(
		template.New("").Funcs(funcMap).ParseGlob("templates/*.html"),
	)
)

func main() {
	b := Blog{}
	posts, err := b.getAllPosts("posts/**/*.md")
	if err != nil {
		log.Fatal(err)
	}
	b.posts = posts

	router := mux.NewRouter()
	router.Use(handlers.ProxyHeaders)
	router.HandleFunc("/favicon.ico", faviconHandler)
	router.HandleFunc("/", b.homeHandler)
	router.NotFoundHandler = http.HandlerFunc(b.homeHandler)
	router.HandleFunc("/{year:20[1-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}/{postName}", b.postHandler)
	router.HandleFunc("/tag/{tagName}", b.tagHandler)
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets", http.FileServer(http.Dir("assets"))))
	router.HandleFunc("/sitemap.xml", b.sitemapHandler)

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

func (b *Blog) getAllPosts(pattern string) ([]*Post, error) {
	files, err := doublestar.Glob(pattern)
	if err != nil {
		return nil, errors.Wrap(err, "doublestar.Glob")
	}

	var posts []*Post

	for _, f := range files {
		post, err := parseMarkdown(f)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse: %s", f)
		}
		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.Time.After(posts[j].Date.Time)
	})

	return posts, nil
}

type publishDate struct {
	time.Time
}

func (d *publishDate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var pd string
	if err := unmarshal(&pd); err != nil {
		return err
	}

	unixDate, err := time.Parse(layoutUnix, pd)
	if err != nil {
		return err
	}

	d.Time = unixDate

	return nil
}

func toISODate(d publishDate) string {
	return d.Time.Format(layoutISO)
}

func getYear(d publishDate) string {
	return strconv.Itoa(d.Time.Year())
}

func getMonth(d publishDate) string {
	month := int(d.Time.Month())
	if month < 10 {
		return "0" + strconv.Itoa(month)
	}

	return strconv.Itoa(month)
}

func getDay(d publishDate) string {
	day := d.Time.Day()
	if day < 10 {
		return "0" + strconv.Itoa(day)
	}

	return strconv.Itoa(day)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

func (b *Blog) homeHandler(w http.ResponseWriter, r *http.Request) {
	if err := renderHTML(w, r, b.posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (b *Blog) postHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	year := vars["year"]
	month := vars["month"]
	fileName := vars["postName"]

	currentPost, err := parseMarkdown(filepath.Join("posts", year, month, fileName+".md"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	relatedPosts, err := getRelatedPosts(b.posts, currentPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	previousPost, nextPost := getPreviousAndNextPost(b.posts, currentPost)
	if previousPost != nil {
		currentPost.HasPrev = true
	}
	if nextPost != nil {
		currentPost.HasNext = true
	}

	data := pongo2.Context{"title": currentPost.Title, "currentPost": currentPost, "relatedPosts": relatedPosts, "previousPost": previousPost, "nextPost": nextPost}
	if err := templates.ExecuteTemplate(w, "post", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func parseMarkdown(filename string) (*Post, error) {
	postContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file: %s", filename)
	}

	var closingMetadataLine int

	lines := strings.Split(string(postContent), "\n")
	for i := 1; i < len(lines); i++ {
		if lines[i] == yamlSeparator {
			closingMetadataLine = i
			break
		}
	}

	metadata := strings.Join(lines[1:closingMetadataLine], "\n")

	p := Post{}
	if err := yaml.Unmarshal([]byte(metadata), &p); err != nil {
		return nil, errors.Wrap(err, "yaml.Unmarshal")
	}
	basename := filepath.Base(filename)
	p.URI = path.Join(getYear(p.Date), getMonth(p.Date), getDay(p.Date), strings.TrimSuffix(basename, filepath.Ext(basename)))

	content := strings.Join(lines[closingMetadataLine+1:], "\n")
	options := []html.Option{
		html.WithLineNumbers(),
	}

	p.Content = template.HTML(bf.Run(
		[]byte(content),
		bf.WithRenderer(
			bfchroma.NewRenderer(
				bfchroma.WithoutAutodetect(),
				bfchroma.ChromaOptions(options...),
				bfchroma.ChromaStyle(styles.SolarizedDark),
			),
		),
	))

	return &p, nil
}

func getRelatedPosts(posts []*Post, currentPost *Post) (map[string]*Post, error) {
	relatedPosts := make(map[string]*Post)
	for _, tag := range currentPost.Tags {
		postsByTag, err := getPostsByTag(posts, tag)
		if err != nil {
			return nil, err
		}

		for _, post := range postsByTag {
			if post.URI != currentPost.URI {
				relatedPosts[post.URI] = post
			}
		}
	}

	return relatedPosts, nil
}

func getPreviousAndNextPost(posts []*Post, currentPost *Post) (previousPost, nextPost *Post) {
	for i, post := range posts {
		if currentPost.URI == post.URI {
			if i < len(posts)-1 {
				previousPost = posts[i+1]
			}
			if i > 0 {
				nextPost = posts[i-1]
			}
			break
		}
	}

	return previousPost, nextPost
}

func (b *Blog) tagHandler(w http.ResponseWriter, r *http.Request) {
	tag := mux.Vars(r)["tagName"]

	postsByTag, err := getPostsByTag(b.posts, tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := renderHTML(w, r, postsByTag); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getPostsByTag(posts []*Post, tag string) ([]*Post, error) {
	var postsByTag []*Post
	for _, post := range posts {
		for _, t := range post.Tags {
			if t == tag {
				postsByTag = append(postsByTag, post)
			}
		}
	}

	return postsByTag, nil
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

type urlset struct {
	XMLNS string `xml:"xmlns,attr"`
	URLs  []url  `xml:"url"`
}

type url struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

func (b *Blog) sitemapHandler(w http.ResponseWriter, r *http.Request) {
	scheme := "http"
	if xForwardedProto := r.Header.Get("X-Forwarded-Proto"); xForwardedProto != "" {
		scheme = xForwardedProto
	}

	sitemap := urlset{
		XMLNS: xmlns,
		URLs: []url{
			{
				Loc: fmt.Sprintf("%s://%s", scheme, r.Host),
			},
		},
	}

	for _, post := range b.posts {
		sitemap.URLs = append(sitemap.URLs, url{
			Loc:     fmt.Sprintf("%s://%s/%s", scheme, r.Host, post.URI),
			LastMod: toISODate(post.Date),
		})
	}

	output, err := xml.MarshalIndent(sitemap, "  ", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(xml.Header + string(output)))
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
