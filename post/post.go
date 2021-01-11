package post

import (
	"context"
	"html/template"
	"io/ioutil"
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
	"github.com/blevesearch/bleve"
	"github.com/pkg/errors"
	bf "github.com/russross/blackfriday/v2"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
)

const (
	yamlSeparator = "---"

	layoutUnix = "Mon Jan 2 15:04:05 -07 2006"
	layoutISO  = "2006-01-02"
)

type Blog interface {
	GetLatestPosts(days int) []*Post
	GetRelatedPosts(currentPost *Post) (map[string]*Post, error)
	GetPostsByTag(tag string) ([]*Post, error)
	GetPreviousAndNextPost(currentPost *Post) (previousPost, nextPost *Post)
	IndexPosts(path string) (bleve.Index, error)
	Search(index bleve.Index, value string) ([]*Post, error)
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

func ToISODate(d publishDate) string {
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

func GetAllPosts(root string) ([]*Post, error) {
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

	postsCh := make(chan *Post)
	g.Go(func() error {
		for f := range paths {
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

	var posts []*Post
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

type blog struct {
	posts []*Post
}

func NewBlog(posts []*Post) *blog {
	return &blog{
		posts: posts,
	}
}

func (b *blog) GetLatestPosts(days int) []*Post {
	var (
		now         = time.Now()
		latestPosts []*Post
	)
	for _, post := range b.posts {
		if post.Date.Time.AddDate(0, 0, days).After(now) {
			latestPosts = append(latestPosts, post)
		} else {
			break
		}
	}

	return latestPosts
}

func ParseMarkdown(ctx context.Context, filename string) (*Post, error) {
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
			return nil, errors.Wrapf(err, "failed to decode metadata of file: %s", filename)
		}
		basename := filepath.Base(filename)
		p.URI = path.Join(getYear(p.Date), getMonth(p.Date), getDay(p.Date), strings.TrimSuffix(basename, filepath.Ext(basename)))

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

func (b *blog) GetRelatedPosts(currentPost *Post) (map[string]*Post, error) {
	relatedPosts := make(map[string]*Post)
	for _, tag := range currentPost.Tags {
		postsByTag, err := b.GetPostsByTag(tag)
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

func (b *blog) GetPostsByTag(tag string) ([]*Post, error) {
	var postsByTag []*Post
	for _, post := range b.posts {
		for _, t := range post.Tags {
			if t == tag {
				postsByTag = append(postsByTag, post)
			}
		}
	}

	return postsByTag, nil
}

func (b *blog) GetPreviousAndNextPost(currentPost *Post) (previousPost, nextPost *Post) {
	for i, post := range b.posts {
		if currentPost.URI == post.URI {
			if i < len(b.posts)-1 {
				previousPost = b.posts[i+1]
			}
			if i > 0 {
				nextPost = b.posts[i-1]
			}
			break
		}
	}

	return previousPost, nextPost
}
