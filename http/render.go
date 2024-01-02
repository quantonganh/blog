package http

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/astaxie/beego/utils/pagination"
	"github.com/pkg/errors"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/pkg/hash"
	"github.com/quantonganh/blog/ui/html"
)

const defaultPostsPerPage = 10

type render struct {
	config      *blog.Config
	postService blog.PostService
}

// NewRender returns new render service
func NewRender(config *blog.Config, postService blog.PostService) blog.Renderer {
	return &render{
		config:      config,
		postService: postService,
	}
}

// RenderPhotos renders photo page
func (r *render) RenderPhotos(w http.ResponseWriter) error {
	tmpl := html.Parse(nil, "photos.html")
	data := map[string]interface{}{
		"categories":     r.postService.GetAllCategories(),
		"imageAddresses": r.postService.GetImageAddresses(),
		"postURIByImage": r.postService.GetPostURIByImage(),
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}

// RenderTags renders tags page
func (r *render) RenderTags(w http.ResponseWriter) error {
	tmpl := html.Parse(nil, "tags.html")
	data := map[string]interface{}{
		"categories":  r.postService.GetAllCategories(),
		"tags":        r.postService.GetAllTags(),
		"postsPerTag": r.postService.GetPostsPerTag(),
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}

// RenderArchives renders archives page
func (r *render) RenderArchives(w http.ResponseWriter) error {
	tmpl := html.Parse(template.FuncMap{
		"toMonthName": blog.ToMonthName,
	}, "archives.html")
	data := map[string]interface{}{
		"categories":   r.postService.GetAllCategories(),
		"years":        r.postService.GetYears(),
		"monthsInYear": r.postService.GetMonthsInYear(),
		"postsByMonth": r.postService.GetPostsByMonth(),
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}

// RenderPosts renders blogs posts
func (r *render) RenderPosts(w http.ResponseWriter, req *http.Request, posts []*blog.Post) error {
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
	paginator := pagination.NewPaginator(req, postsPerPage, int64(nums))
	offset := paginator.Offset()

	endPos := offset + postsPerPage
	if endPos > nums {
		endPos = nums
	}

	tmpl := html.Parse(template.FuncMap{
		"toISODate": blog.ToISODate,
	}, "home.html")
	data := map[string]interface{}{
		"Site":       r.config.Site,
		"categories": r.postService.GetAllCategories(),
		"posts":      posts[offset:endPos],
		"paginator":  paginator,
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}

// RenderPost renders a single blog post
func (r *render) RenderPost(w http.ResponseWriter, currentPost *blog.Post, relatedPosts []*blog.Post, previousPost, nextPost *blog.Post) error {
	tmpl := html.Parse(template.FuncMap{
		"toISODate": blog.ToISODate,
	}, "post.html")
	data := map[string]interface{}{
		"categories":   r.postService.GetAllCategories(),
		"Title":        currentPost.Title,
		"Description":  currentPost.Description,
		"currentPost":  currentPost,
		"relatedPosts": relatedPosts,
		"previousPost": previousPost,
		"nextPost":     nextPost,
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		return errors.Errorf("failed to render post %s: %v", currentPost.Title, err)
	}

	return nil
}

// RenderResponseMessage renders HTTP response message
func (r *render) RenderResponseMessage(w http.ResponseWriter, contextualClass, message string) error {
	tmpl := html.Parse(nil, "subscribe.html")
	data := map[string]interface{}{
		"categories":      r.postService.GetAllCategories(),
		"contextualClass": contextualClass,
		"message":         message,
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}

// RenderNewsletter renders newsletter
func (r *render) RenderNewsletter(latestPosts []*blog.Post, serverURL, email string) (*bytes.Buffer, error) {
	funcMap := template.FuncMap{
		"toISODate": blog.ToISODate,
	}
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(html.FS, "newsletter.html"))
	hash, err := hash.ComputeHmac256(email, r.config.Newsletter.HMAC.Secret)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	data := map[string]interface{}{
		"categories": r.postService.GetAllCategories(),
		"posts":      latestPosts,
		"pageURL":    serverURL,
		"email":      email,
		"hash":       hash,
	}
	if err := tmpl.ExecuteTemplate(buf, "newsletter", data); err != nil {
		return nil, errors.Errorf("failed to execute template newsletter: %v", err)
	}

	return buf, nil
}

func (r *render) RenderVTV(w http.ResponseWriter, letters string, total int, rows [][]string) error {
	tmpl := html.Parse(template.FuncMap{
		"mod": func(i, j, r int) bool {
			return i%j == r
		},
	}, "vtv.html")
	data := map[string]interface{}{
		"letters": letters,
		"total":   total,
		"rows":    rows,
	}
	return tmpl.ExecuteTemplate(w, "base", data)
}
