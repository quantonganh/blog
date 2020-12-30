package main

import (
	"context"
	"html/template"
	"io/ioutil"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Depado/bfchroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/styles"
	"github.com/bmatcuk/doublestar/v2"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	bf "gopkg.in/russross/blackfriday.v2"
	"gopkg.in/yaml.v2"
)

const yamlSeparator = "---"

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