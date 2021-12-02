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
	newLineSeparator     = "\n"
	yamlSeparator        = "---"
	travelCategory       = "Du lịch"
	wordSeparator        = " "
	summaryLength        = 70
	threeBackticks       = "```"
	numberOfRelatedPosts = 5
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
			post, err := ParseMarkdown(ctx, root, f)
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

func (ps *postService) GetAllPosts() []*blog.Post {
	return ps.posts
}

func (ps *postService) GetPostByURI(uri string) *blog.Post {
	for _, post := range ps.posts {
		if post.URI == uri {
			return post
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

func ParseMarkdown(ctx context.Context, root string, r io.Reader) (*blog.Post, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		postContent, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read from io.Reader")
		}

		var closingMetadataLine int

		lines := strings.Split(string(postContent), newLineSeparator)
		for i := 1; i < len(lines); i++ {
			if lines[i] == yamlSeparator {
				closingMetadataLine = i
				break
			}
		}

		metadata := strings.Join(lines[1:closingMetadataLine], newLineSeparator)

		p := blog.Post{}
		if err := yaml.Unmarshal([]byte(metadata), &p); err != nil {
			return nil, errors.Wrapf(err, "failed to decode metadata")
		}

		switch v := r.(type) {
		case *os.File:
			name := v.Name()
			basename := filepath.Base(name)
			p.URI = path.Join(strings.TrimPrefix(filepath.Dir(name), root), strings.TrimSuffix(basename, filepath.Ext(basename)))
		default:
			p.URI = path.Join(p.Date.GetYear(), p.Date.GetMonth(), p.Date.GetDay(), url.QueryEscape(strings.ToLower(p.Title)))
		}
		if !strings.HasPrefix(p.URI, "/") {
			p.URI = "/" + p.URI
		}

		content := strings.Join(lines[closingMetadataLine+1:], newLineSeparator)
		if len(strings.Split(content, wordSeparator)) > summaryLength {
			p.Truncated = true
		} else {
			p.Truncated = false
		}
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

		var (
			summaries         []string
			numThreeBackticks int
		)
		for _, line := range lines[closingMetadataLine+1:] {
			if strings.Contains(line, threeBackticks) {
				numThreeBackticks++
			}
			summaries = append(summaries, line)
			if len(strings.Split(strings.Join(summaries, newLineSeparator), wordSeparator)) > summaryLength && numThreeBackticks%2 == 0 {
				break
			}
		}
		p.Summary = template.HTML(bf.Run(
			[]byte(strings.Join(summaries, newLineSeparator)),
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

func (ps *postService) GetRelatedPosts(currentPost *blog.Post) []*blog.Post {
	var (
		m            = make(map[int]*blog.Post)
		relatedPosts []*blog.Post
	)
	for _, tag := range currentPost.Tags {
		for _, post := range ps.posts {
			if post.ID != currentPost.ID && contains(post.Tags, tag) {
				_, found := m[post.ID]
				m[post.ID] = post
				if !found {
					relatedPosts = append(relatedPosts, post)
				}
			}

			if len(relatedPosts) >= numberOfRelatedPosts {
				return relatedPosts[:numberOfRelatedPosts]
			}
		}
	}

	for _, category := range currentPost.Categories {
		for _, post := range ps.posts {
			if post.ID != currentPost.ID && contains(post.Categories, category) {
				_, found := m[post.ID]
				m[post.ID] = post
				if !found {
					relatedPosts = append(relatedPosts, post)
				}
			}

			if len(relatedPosts) >= numberOfRelatedPosts {
				return relatedPosts[:numberOfRelatedPosts]
			}
		}
	}

	return relatedPosts
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func (ps *postService) GetAllCategories() map[string][]*blog.Post {
	categories := make(map[string][]*blog.Post)
	for _, post := range ps.posts {
		for _, c := range post.Categories {
			categories[c] = append(categories[c], post)
		}
	}

	return categories
}

func (ps *postService) GetPostsPerTag() map[string]int {
	m := make(map[string]int)
	for _, post := range ps.posts {
		for _, t := range post.Tags {
			m[t]++
		}
	}

	return m
}

func (ps *postService) GetAllTags() []string {
	m := ps.GetPostsPerTag()

	type kv struct {
		tag   string
		count int
	}

	var tc []kv
	for t, c := range m {
		tc = append(tc, kv{
			tag:   t,
			count: c,
		})
	}
	sort.Slice(tc, func(i, j int) bool {
		return tc[i].count > tc[j].count
	})

	tags := make([]string, 0, len(m))
	for _, kv := range tc {
		tags = append(tags, kv.tag)
	}

	return tags
}

func (ps *postService) GetImageAddresses() []string {
	var imageAddresses []string
	for _, post := range ps.posts {
		if blog.Contains(post.Categories, travelCategory) {
			imageAddresses = append(imageAddresses, post.Images...)
		}
	}
	return imageAddresses
}

func (ps *postService) GetPostURIByImage() map[string]string {
	m := make(map[string]string)
	for _, post := range ps.posts {
		if blog.Contains(post.Categories, travelCategory) {
			for _, image := range post.Images {
				m[image] = post.URI
			}
		}
	}
	return m
}

func (ps *postService) GetPostsByCategory(category string) []*blog.Post {
	return ps.GetAllCategories()[category]
}

func (ps *postService) GetPostsByTag(tag string) []*blog.Post {
	var postsByTag []*blog.Post
	for _, post := range ps.posts {
		for _, t := range post.Tags {
			if t == tag {
				postsByTag = append(postsByTag, post)
			}
		}
	}

	return postsByTag
}

func (ps *postService) GetPreviousAndNextPost(currentPost *blog.Post) (previousPost, nextPost *blog.Post) {
	id := currentPost.ID
	if id < len(ps.posts)-1 {
		previousPost = ps.posts[id+1]
	}
	if id > 0 {
		nextPost = ps.posts[id-1]
	}

	return previousPost, nextPost
}

func (ps *postService) GetYears() []string {
	m := make(map[string]struct{})
	for _, post := range ps.posts {
		m[post.Date.GetYear()] = struct{}{}
	}

	var years []string
	for y := range m {
		years = append(years, y)
	}
	sort.Slice(years, func(i, j int) bool {
		return years[i] > years[j]
	})
	return years
}

func (ps *postService) GetMonthsInYear() map[string][]string {
	monthsInYear := make(map[string][]string)
	for _, year := range ps.GetYears() {
		var (
			m      = make(map[string]struct{})
			months []string
		)
		for _, post := range ps.posts {
			if post.Date.GetYear() == year {
				month := post.Date.GetMonth()
				_, found := m[month]
				m[month] = struct{}{}
				if !found {
					months = append(months, month)
				}
			}
		}

		sort.Slice(months, func(i, j int) bool {
			return months[i] > months[j]
		})
		monthsInYear[year] = months
	}

	return monthsInYear
}

func (ps *postService) GetPostsByDate(year, month, day string) []*blog.Post {
	var postsByDate []*blog.Post
	for _, post := range ps.posts {
		if post.Date.GetYear() == year && post.Date.GetMonth() == month && post.Date.GetDay() == day {
			postsByDate = append(postsByDate, post)
		}
	}
	return postsByDate
}

func (ps *postService) GetPostsByMonth() map[string]map[string][]*blog.Post {
	ym := make(map[string]map[string][]*blog.Post)
	for _, year := range ps.GetYears() {
		mm := make(map[string][]*blog.Post)
		for _, month := range ps.GetMonthsInYear()[year] {
			for _, post := range ps.posts {
				if post.Date.GetYear() == year && post.Date.GetMonth() == month {
					mm[month] = append(mm[month], post)
				}
			}
		}
		ym[year] = mm
	}
	return ym
}

func (ps *postService) GetPostsByYear(year string) []*blog.Post {
	var postsByYear []*blog.Post
	for _, post := range ps.posts {
		if post.Date.GetYear() == year {
			postsByYear = append(postsByYear, post)
		}
	}
	return postsByYear
}
