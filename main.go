package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/index/scorch"
	"github.com/blevesearch/bleve/mapping"
	"github.com/bmatcuk/doublestar/v2"
	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
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
	router.HandleFunc("/sitemap.xml", mwError(b.sitemapHandler))
	router.HandleFunc("/search", mwError(b.searchHandler))

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

func getAllPosts(pattern string) ([]*Post, error) {
	files, err := doublestar.Glob(pattern)
	if err != nil {
		return nil, errors.Wrap(err, "doublestar.Glob")
	}

	g, ctx := errgroup.WithContext(context.Background())
	postsCh := make(chan *Post)
	for _, f := range files {
		f := f // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			post, err := parseMarkdown(ctx, f)
			if err != nil {
				return err
			}
			select {
			case postsCh <- post:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}

	go func() {
		_ = g.Wait()
		close(postsCh)
	}()

	var posts []*Post
	for post := range postsCh {
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

	layouts := []string{layoutUnix, layoutISO}
	for _, layout := range layouts {
		date, err := time.Parse(layout, pd)
		if err == nil {
			d.Time = date
			return nil
		}
	}

	return errors.Errorf("Unrecognized date format: %s", pd)
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

type remark struct {
	URL     *url.URL
	PageURL *url.URL
}

func getRemarkURL() (*remark, error) {
	remarkURL, err := url.Parse(os.Getenv("REMARK_URL"))
	if err != nil {
		return nil, err
	}
	pageURL, err := url.Parse(os.Getenv("PAGE_URL"))
	if err != nil {
		return nil, err
	}

	return &remark{
		URL:     remarkURL,
		PageURL: pageURL,
	}, nil
}

func parseMarkdown(ctx context.Context, filename string) (*Post, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
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

func (b *Blog) indexPosts(path string) (bleve.Index, error) {
	indexMapping := bleve.NewIndexMapping()
	index, err := bleve.NewUsing(path, indexMapping, scorch.Name, scorch.Name, nil)
	if err != nil {
		return nil, errors.Errorf("failed to create index at %s: %v", path, err)
	}

	g, ctx := errgroup.WithContext(context.Background())
	for _, post := range b.posts {
		post := post // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			return indexPost(ctx, indexMapping, index, post)
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return index, nil
}

func indexPost(ctx context.Context, mapping *mapping.IndexMappingImpl, index bleve.Index, post *Post) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		doc := document.Document{
			ID: post.URI,
		}
		if err := mapping.MapDocument(&doc, post); err != nil {
			return errors.Errorf("failed to map document: %v", err)
		}

		var b bytes.Buffer
		enc := gob.NewEncoder(&b)
		if err := enc.Encode(post); err != nil {
			return errors.Errorf("failed to encode post: %v", err)
		}

		field := document.NewTextFieldWithIndexingOptions("_source", nil, b.Bytes(), document.StoreField)
		batch := index.NewBatch()
		if err := batch.IndexAdvanced(doc.AddField(field)); err != nil {
			return errors.Errorf("failed to add index to the batch: %v", err)
		}
		if err := index.Batch(batch); err != nil {
			return errors.Errorf("failed to index batch: %v", err)
		}

		return nil
	}
}

func (b *Blog) search(index bleve.Index, value string) ([]*Post, error) {
	query := bleve.NewMatchQuery(value)
	request := bleve.NewSearchRequest(query)
	request.Fields = []string{"_source"}
	searchResults, err := index.Search(request)
	if err != nil {
		return nil, errors.Errorf("failed to execute a search request: %v", err)
	}

	var searchPosts []*Post
	for _, result := range searchResults.Hits {
		var post *Post
		b := bytes.NewBuffer([]byte(fmt.Sprintf("%v", result.Fields["_source"])))
		dec := gob.NewDecoder(b)
		if err = dec.Decode(&post); err != nil {
			return nil, errors.Errorf("failed to decode post: %v", err)
		}
		searchPosts = append(searchPosts, post)
	}

	return searchPosts, nil
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
