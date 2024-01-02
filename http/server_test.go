package http

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	nethtml "golang.org/x/net/html"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/markdown"
)

var (
	cfg  *blog.Config
	post *blog.Post
	s    *Server
)

func TestMain(m *testing.M) {
	viper.SetConfigType("yaml")
	var yamlConfig = []byte(`
posts:
  dir: test

templates:
  dir: html/templates

newsletter:
  hmac:
    secret: da02e221bc331c9875c5e1299fa8d765
`)
	if err := viper.ReadConfig(bytes.NewBuffer(yamlConfig)); err != nil {
		log.Fatal(err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}

	r := strings.NewReader(`---
title: Test
date: Thu Sep 19 21:48:39 +07 2019
description: Just a test
images:
  - /path/to/photo.jpg
categories:
  - Du lịch
tags:
  - test
---
Test.`)
	var err error
	post, err = markdown.Parse(context.Background(), ".", r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("post: %+v", post)

	s, err = NewServer(cfg, []*blog.Post{post})
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestParseMarkdown(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Test", post.Title)
	assert.Equal(t, "2019-09-19", blog.ToISODate(post.Date))
	assert.Equal(t, "Just a test", post.Description)
	assert.Equal(t, "test", post.Tags[0])
	assert.Equal(t, template.HTML("<p>Test.</p>\n"), post.Content)
}

func TestFaviconHandler(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/favicon.ico", nil)
	assert.NoError(t, err)
	s.router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHomeHandler(t *testing.T) {
	t.Run("home", func(t *testing.T) {
		testPostHandler(t, "/")
	})
}

func TestArchivesHandler(t *testing.T) {
	t.Run("archives", func(t *testing.T) {
		testPostHandler(t, "/archives")
	})
}

func TestPostHandler(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/2019/09/19/test", nil)
	assert.NoError(t, err)
	s.router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "/tags/test", getLinkByText(t, rr.Body, "test"))
}

func TestTagHandler(t *testing.T) {
	t.Run("tags", func(t *testing.T) {
		testPostHandler(t, "/tags/test")
	})
}

func TestSearchHandler(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	formData := url.Values{}
	formData.Add("q", "test")
	request, err := http.NewRequest(http.MethodPost, "/search", strings.NewReader(formData.Encode()))
	assert.NoError(t, err)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s.router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "/2019/09/19/test.md", getLinkByText(t, rr.Body, "Test"))

	t.Cleanup(func() {
		_ = os.RemoveAll("test.bleve")
	})
}

func getLinkByText(t *testing.T, body *bytes.Buffer, text string) string {
	doc, err := goquery.NewDocumentFromReader(body)
	require.NoError(t, err)

	var link string
	doc.Find("article a").Each(func(_ int, s *goquery.Selection) {
		if strings.TrimSpace(s.Text()) == text {
			link, _ = s.Attr("href")
		}
	})

	return link
}

func getResponseMessage(body io.ReadCloser) (string, error) {
	tokenizer := nethtml.NewTokenizer(body)
	inDiv := false
	var buffer bytes.Buffer

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			if err := tokenizer.Err(); err == io.EOF {
				break
			}
			return "", nil
		}

		token := tokenizer.Token()

		switch tokenType {
		case nethtml.StartTagToken:
			if token.Data == "div" {
				// Check if the div has the expected class attribute
				for _, attr := range token.Attr {
					fmt.Printf("attr: %+v\n", attr)
					if attr.Key == "class" && strings.HasPrefix(attr.Val, "alert") {
						inDiv = true
						break
					}
				}
			}
		case nethtml.TextToken:
			if inDiv {
				buffer.WriteString(token.Data)
			}
		case nethtml.EndTagToken:
			if inDiv && token.Data == "div" {
				return buffer.String(), nil
			}
		}
	}

	return "", nil
}
