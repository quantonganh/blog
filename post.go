package blog

import (
	"html/template"
	"path"
	"sort"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	layoutUnix = "Mon Jan 2 15:04:05 -07 2006"
	layoutISO  = "2006-01-02"
)

type PostService interface {
	GetAllPosts() []*Post
	GetPostByURI(uri string) *Post
	GetLatestPosts(days int) []*Post
	GetRelatedPosts(currentPost *Post) []*Post
	GetAllCategories() map[string][]*Post
	GetAllTags() []string
	GetImageAddresses() []string
	GetPostsByCategory(category string) []*Post
	GetPostsByTag(tag string) []*Post
	GetPreviousAndNextPost(currentPost *Post) (previousPost, nextPost *Post)
	GetYears() []string
	GetPostsByDate(year, month, date string) []*Post
	GetPostsByMonth(year, month string) []*Post
	GetPostsByYear(year string) []*Post
	Search(value string) ([]*Post, error)
	CloseIndex() error
}

type Post struct {
	ID          int
	URI         string
	Title       string
	Date        publishDate
	Description string
	Images      []string
	Content     template.HTML
	Summary     template.HTML
	Truncated   bool
	Categories  []string
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

func CountPostsOnATag(posts []*Post, tag string) int {
	c := 0
	for _, post := range posts {
		if Contains(post.Tags, tag) {
			c++
		}
	}

	return c
}

func ToMonthName(month string) string {
	m, _ := strconv.Atoi(month)
	return time.Month(m).String()
}

func GetMonthsInYear(posts []*Post, year string) []string {
	m := make(map[string]struct{})
	for _, post := range posts {
		if post.Date.GetYear() == year {
			m[post.Date.GetMonth()] = struct{}{}
		}
	}

	var monthsInYear []string
	for month := range m {
		monthsInYear = append(monthsInYear, month)
	}
	sort.Slice(monthsInYear, func(i, j int) bool {
		return monthsInYear[i] > monthsInYear[j]
	})

	return monthsInYear
}

func GetPostsByMonth(posts []*Post, year, month string) []*Post {
	var postsByMonth []*Post
	for _, post := range posts {
		if post.Date.GetYear() == year && post.Date.GetMonth() == month {
			postsByMonth = append(postsByMonth, post)
		}
	}
	return postsByMonth
}

func GetPostURIByImage(posts []*Post, imageAddress string) string {
	for _, post := range posts {
		if Contains(post.Images, path.Base(imageAddress)) {
			return post.URI
		}
	}
	return ""
}

func Contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func (d *publishDate) GetYear() string {
	return strconv.Itoa(d.Time.Year())
}

func (d *publishDate) GetMonth() string {
	month := int(d.Time.Month())
	if month < 10 {
		return "0" + strconv.Itoa(month)
	}

	return strconv.Itoa(month)
}

func (d *publishDate) GetDay() string {
	day := d.Time.Day()
	if day < 10 {
		return "0" + strconv.Itoa(day)
	}

	return strconv.Itoa(day)
}
