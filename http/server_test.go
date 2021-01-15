package http

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/http/html"
	"github.com/quantonganh/blog/http/mw"
	"github.com/quantonganh/blog/ondisk"
)

func TestHandler(t *testing.T) {
	viper.SetConfigType("yaml")
	var yamlConfig = []byte(`
newsletter:
  hmac:
    secret: da02e221bc331c9875c5e1299fa8d765
`)
	require.NoError(t, viper.ReadConfig(bytes.NewBuffer(yamlConfig)))
	var cfg *blog.Config
	require.NoError(t, viper.Unmarshal(&cfg))

	r := strings.NewReader(`---
title: Test
date: Thu Sep 19 21:48:39 +07 2019
description: Just a test
tags:
  - test
---
Test.`)
	testPost, err := ondisk.ParseMarkdown(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, "Test", testPost.Title)
	assert.Equal(t, "2019-09-19", blog.ToISODate(testPost.Date))
	assert.Equal(t, "Just a test", testPost.Description)
	assert.Equal(t, "test", testPost.Tags[0])
	assert.Equal(t, template.HTML("<p>Test.</p>\n"), testPost.Content)

	posts := []*blog.Post{testPost}
	funcMap := template.FuncMap{
		"toISODate": blog.ToISODate,
	}
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob("html/templates/*.tmpl"))
	s := &Server{
		PostService: ondisk.NewPostService(posts),
		Renderer:    html.NewRender(cfg, tmpl),
	}
	t.Run("homeHandler", s.testHomeHandler)

	t.Run("postHandler", s.testPostHandler)

	t.Run("tagHandler", s.testTagHandler)

	t.Run("searchHandler", s.testSearchHandler)

	t.Run("subscribeHandler", s.testSubscribeHandler)

	t.Run("confirmHandler", s.testConfirmHandler)

	t.Run("unsubscribeHandler", func(t *testing.T) {
		s.testUnsubscribeHandler(t, cfg.Newsletter.HMAC.Secret)
	})
}

func (s *Server) testHomeHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/", mw.Error(s.homeHandler))

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "/2019/09/19/test", getLinkByText(t, rr.Body, "Test"))
}

func (s *Server) testPostHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/{year:20[1-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}/{postName}", mw.Error(s.postHandler))

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/2019/09/19/test", nil)
	assert.NoError(t, err)
	router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "/tag/test", getLinkByText(t, rr.Body, "#test"))
}

func (s *Server) testTagHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/tag/{tagName}", mw.Error(s.tagHandler))

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/tag/test", nil)
	assert.NoError(t, err)
	router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "/2019/09/19/test", getLinkByText(t, rr.Body, "Test"))
}

func (s *Server) testSearchHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/search", mw.Error(s.searchHandler("test.bleve")))

	rr := httptest.NewRecorder()
	formData := url.Values{}
	formData.Add("q", "test")
	request, err := http.NewRequest(http.MethodPost, "/search", strings.NewReader(formData.Encode()))
	assert.NoError(t, err)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "/2019/09/19/test", getLinkByText(t, rr.Body, "Test"))

	t.Cleanup(func() {
		_ = os.RemoveAll("test.bleve")
	})
}

func getLinkByText(t *testing.T, body *bytes.Buffer, text string) string {
	doc, err := goquery.NewDocumentFromReader(body)
	require.NoError(t, err)

	var link string
	doc.Find("article a").Each(func(_ int, s *goquery.Selection) {
		if s.Text() == text {
			link, _ = s.Attr("href")
		}
	})

	return link
}
