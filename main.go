package main

import (
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
	"github.com/flosch/pongo2"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	bf "gopkg.in/russross/blackfriday.v2"
	yaml "gopkg.in/yaml.v2"
)

const (
	yamlDelim  = "---"
	unixLayout = "Mon Jan 2 15:04:05 -07 2006"
	postLayout = "2-Jan-2006"
)

var (
	funcMap   = template.FuncMap{"formatDate": formatDate}
	templates = template.Must(
		template.New("").Funcs(funcMap).ParseFiles(
			"templates/header.html",
			"templates/footer.html",
			"templates/paginator.html",
			"templates/home.html",
			"templates/posts.html",
			"templates/post.html",
		),
	)
)

type Post struct {
	Title       string
	Date        publishDate
	Description string
	Content     template.HTML
	Tags        []string
	File        string
}

type publishDate struct {
	time.Time
}

func (d *publishDate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var pd string
	if err := unmarshal(&pd); err != nil {
		return err
	}

	unixDate, err := time.Parse(unixLayout, pd)
	if err != nil {
		return err
	}
	d.Time = unixDate

	return nil
}

func formatDate(d publishDate) string {
	return d.Time.Format(postLayout)
}

func homeHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	posts, err := listAllPosts("posts/*.md")
	if err != nil {
		log.Fatal(err)
	}

	if err := renderHTML(w, r, posts); err != nil {
		log.Fatal(err)
	}
}

func tagHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tag := p.ByName("tagName")

	posts, err := listAllPosts("posts/*.md")
	if err != nil {
		log.Fatal(err)
	}
	postsByTag := []*Post{}
	for _, post := range posts {
		for _, t := range post.Tags {
			if t == tag {
				postsByTag = append(postsByTag, post)
			}
		}
	}

	if err := renderHTML(w, r, postsByTag); err != nil {
		log.Fatal(err)
	}
}

func renderHTML(w http.ResponseWriter, r *http.Request, posts []*Post) error {
	postsPerPageEnv, exists := os.LookupEnv("POSTS_PER_PAGE")
	if !exists {
		postsPerPageEnv = "10"
	}
	postsPerPage, err := strconv.Atoi(postsPerPageEnv)
	if err != nil {
		log.Fatal(err)
	}

	nums := len(posts)
	paginator := pagination.NewPaginator(r, postsPerPage, int64(nums))
	offset := paginator.Offset()
	endPos := offset + postsPerPage
	if endPos > nums {
		endPos = nums
	}
	data := pongo2.Context{"paginator": paginator, "posts": posts[offset:endPos]}
	if err := templates.ExecuteTemplate(w, "home", data); err != nil {
		log.Fatal(err)
	}

	return nil
}

func postsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	posts, err := listAllPosts("posts/*.md")
	if err != nil {
		log.Fatal(err)
	}

	if err := templates.ExecuteTemplate(w, "posts", &posts); err != nil {
		log.Fatal(err)
	}
}

func listAllPosts(pattern string) ([]*Post, error) {
	posts := []*Post{}
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, errors.Wrap(err, "filepath.Glob")
	}

	for _, f := range files {
		post, err := parseMarkdown(f)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse: %s", f)
		}
		filename := filepath.Base(f)
		post.File = strings.TrimSuffix(filename, path.Ext(filename))
		posts = append(posts, post)
	}
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.Time.After(posts[j].Date.Time)
	})

	return posts, nil
}

func postHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	path := p.ByName("postName")
	post, err := parseMarkdown("posts/" + path + ".md")
	if err != nil {
		log.Fatal(err)
	}
	post.File = path

	if err := templates.ExecuteTemplate(w, "post", &post); err != nil {
		log.Fatal(err)
	}
}

func parseMarkdown(f string) (*Post, error) {
	fileread, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, errors.Wrap(err, "ioutil.ReadFile")
	}

	lines := strings.Split(string(fileread), "\n")
	var closingMetadataLine int
	for i := 1; i < len(lines); i++ {
		if lines[i] == yamlDelim {
			closingMetadataLine = i
		}
	}
	metadata := strings.Join(lines[1:closingMetadataLine], "\n")

	p := Post{}
	if err := yaml.Unmarshal([]byte(metadata), &p); err != nil {
		return nil, errors.Wrap(err, "yaml.Unmarshal")
	}
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

func faviconHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "favicon.ico")
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

func main() {
	router := httprouter.New()
	router.GET("/favicon.ico", faviconHandler)
	router.GET("/", homeHandler)
	router.GET("/posts", postsHandler)
	router.GET("/posts/:postName", postHandler)
	router.GET("/tags/:tagName", tagHandler)
	router.ServeFiles("/assets/*filepath", http.Dir("assets"))

	log.Fatal(http.ListenAndServe(":80", logHandler(router)))
}
