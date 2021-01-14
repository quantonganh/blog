package ondisk

import (
	"context"
	"html/template"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Depado/bfchroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/styles"
	"github.com/pkg/errors"
	bf "github.com/russross/blackfriday/v2"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"

	"github.com/quantonganh/blog"
)

const (
	yamlSeparator = "---"
)

func GetAllPosts(root string) ([]*blog.Post, error) {
	g, ctx := errgroup.WithContext(context.Background())
	paths := make(chan string)
	g.Go(func() error {
		defer close(paths)
		return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(path) != ".md" {
				return nil
			}
			select {
			case paths <- path:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	})

	postsCh := make(chan *blog.Post)
	g.Go(func() error {
		for p := range paths {
			f, err := os.Open(p)
			if err != nil {
				return errors.Wrapf(err, "failed to open file: %s", p)
			}
			post, err := ParseMarkdown(ctx, f)
			if err != nil {
				return err
			}
			select {
			case postsCh <- post:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	go func() {
		_ = g.Wait()
		close(postsCh)
	}()

	var posts []*blog.Post
	for post := range postsCh {
		posts = append(posts, post)
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.Time.After(posts[j].Date.Time)
	})

	return posts, nil
}

type postService struct {
	posts []*blog.Post
}

func NewPostService(posts []*blog.Post) *postService {
	return &postService{
		posts: posts,
	}
}

func (ps *postService) GetPostByURI(uri string) *blog.Post {
	for _, p := range ps.posts {
		if p.URI == uri {
			return p
		}
	}

	return nil
}

func (ps *postService) GetLatestPosts(days int) []*blog.Post {
	var (
		now         = time.Now()
		latestPosts []*blog.Post
	)
	for _, post := range ps.posts {
		if post.Date.Time.AddDate(0, 0, days).After(now) {
			latestPosts = append(latestPosts, post)
		} else {
			break
		}
	}

	return latestPosts
}

func ParseMarkdown(ctx context.Context, r io.Reader) (*blog.Post, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		postContent, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read from io.Reader")
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

		p := blog.Post{}
		if err := yaml.Unmarshal([]byte(metadata), &p); err != nil {
			return nil, errors.Wrapf(err, "failed to decode metadata")
		}
		switch v := r.(type) {
		case *os.File:
			basename := filepath.Base(v.Name())
			p.URI = path.Join(blog.GetYear(p.Date), blog.GetMonth(p.Date), blog.GetDay(p.Date), strings.TrimSuffix(basename, filepath.Ext(basename)))
		default:
			p.URI = path.Join(blog.GetYear(p.Date), blog.GetMonth(p.Date), blog.GetDay(p.Date), url.QueryEscape(strings.ToLower(p.Title)))
		}

		content := strings.Join(lines[closingMetadataLine+1:], "\n")
		options := []html.Option{
			html.WithLineNumbers(true),
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

func (ps *postService) GetRelatedPosts(currentPost *blog.Post) (map[string]*blog.Post, error) {
	relatedPosts := make(map[string]*blog.Post)
	for _, tag := range currentPost.Tags {
		postsByTag, err := ps.GetPostsByTag(tag)
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

func (ps *postService) GetPostsByTag(tag string) ([]*blog.Post, error) {
	var postsByTag []*blog.Post
	for _, post := range ps.posts {
		for _, t := range post.Tags {
			if t == tag {
				postsByTag = append(postsByTag, post)
			}
		}
	}

	return postsByTag, nil
}

func (ps *postService) GetPreviousAndNextPost(currentPost *blog.Post) (previousPost, nextPost *blog.Post) {
	for i, post := range ps.posts {
		if currentPost.URI == post.URI {
			if i < len(ps.posts)-1 {
				previousPost = ps.posts[i+1]
			}
			if i > 0 {
				nextPost = ps.posts[i-1]
			}
			break
		}
	}

	return previousPost, nextPost
}
