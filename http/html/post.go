package html

import (
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/astaxie/beego/utils/pagination"
	"github.com/flosch/pongo2"
	"github.com/pkg/errors"

	"github.com/quantonganh/blog"
)

const defaultPostsPerPage = 10

type post struct {
	tmpl *template.Template
}

func NewPost(tmpl *template.Template) *post {
	return &post{
		tmpl: tmpl,
	}
}

func (p *post) Render(w http.ResponseWriter, r *http.Request, posts []*blog.Post) error {
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
	if err := p.tmpl.ExecuteTemplate(w, "home", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}
