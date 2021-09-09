package html

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/astaxie/beego/utils/pagination"
	"github.com/flosch/pongo2"
	"github.com/pkg/errors"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/pkg/hash"
)

const defaultPostsPerPage = 10

type render struct {
	config      *blog.Config
	postService blog.PostService
	tmpl        *template.Template
}

func NewRender(config *blog.Config, postService blog.PostService, tmpl *template.Template) *render {
	return &render{
		config:      config,
		postService: postService,
		tmpl:        tmpl,
	}
}

func (r *render) RenderPhotos(w http.ResponseWriter) error {
	data := pongo2.Context{
		"allPosts":       r.postService.GetAllPosts(),
		"categories":     r.postService.GetAllCategories(),
		"imageAddresses": r.postService.GetImageAddresses(),
	}
	if err := r.tmpl.ExecuteTemplate(w, "photos", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}

func (r *render) RenderArchives(w http.ResponseWriter) error {
	data := pongo2.Context{
		"allPosts":   r.postService.GetAllPosts(),
		"categories": r.postService.GetAllCategories(),
		"years":      r.postService.GetYears(),
	}
	if err := r.tmpl.ExecuteTemplate(w, "archives", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}

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

	data := pongo2.Context{
		"Site":       r.config.Site,
		"categories": r.postService.GetAllCategories(),
		"posts":      posts[offset:endPos],
		"paginator":  paginator,
	}
	if err := r.tmpl.ExecuteTemplate(w, "home", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}

func (r *render) RenderPost(w http.ResponseWriter, currentPost *blog.Post, relatedPosts []*blog.Post, previousPost, nextPost *blog.Post) error {
	data := pongo2.Context{
		"categories":   r.postService.GetAllCategories(),
		"Title":        currentPost.Title,
		"Description":  currentPost.Description,
		"currentPost":  currentPost,
		"relatedPosts": relatedPosts,
		"previousPost": previousPost,
		"nextPost":     nextPost,
	}
	if err := r.tmpl.ExecuteTemplate(w, "post", data); err != nil {
		return errors.Errorf("failed to render post %s: %v", currentPost.Title, err)
	}

	return nil
}

func (r *render) RenderResponseMessage(w http.ResponseWriter, message string) error {
	data := pongo2.Context{
		"categories": r.postService.GetAllCategories(),
		"message":    message,
	}

	if err := r.tmpl.ExecuteTemplate(w, "response", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}

func (r *render) RenderNewsletter(latestPosts []*blog.Post, serverURL, email string) (*bytes.Buffer, error) {
	hash, err := hash.ComputeHmac256(email, r.config.Newsletter.HMAC.Secret)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	data := pongo2.Context{
		"categories": r.postService.GetAllCategories(),
		"posts":      latestPosts,
		"pageURL":    serverURL,
		"email":      email,
		"hash":       hash,
	}
	if err := r.tmpl.ExecuteTemplate(buf, "newsletter", data); err != nil {
		return nil, errors.Errorf("failed to execute template newsletter: %v", err)
	}

	return buf, nil
}
