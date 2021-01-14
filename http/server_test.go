package http

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/http/mw"
	"github.com/quantonganh/blog/ondisk"
)

func TestHandler(t *testing.T) {
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
	s := &Server{
		Templates:   template.Must(template.New("").Funcs(funcMap).ParseGlob("html/templates/*.tmpl")),
		PostService: ondisk.NewPostService(posts),
	}
	t.Run("homeHandler", func(t *testing.T) {
		s.testHomeHandler(t, posts)
	})

	t.Run("postHandler", s.testPostHandler)

	t.Run("tagHandler", s.testTagHandler)

	t.Run("searchHandler", s.testSearchHandler)

	t.Run("subscribeHandler", s.testSubscribeHandler)

	t.Run("confirmHandler", s.testConfirmHandler)

	t.Run("unsubscribeHandler", s.testUnsubscribeHandler)
}

func (s *Server) testHomeHandler(t *testing.T, posts []*blog.Post) {
	router := mux.NewRouter()
	router.HandleFunc("/", mw.Error(s.homeHandler(posts)))

	writer := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
}

func (s *Server) testPostHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/{year:20[1-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}/{postName}", mw.Error(s.postHandler))

	writer := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/2019/09/19/test", nil)
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
}

func (s *Server) testTagHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/tag/test", mw.Error(s.tagHandler))

	writer := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/tag/test", nil)
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
}

func (s *Server) testSearchHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/search", mw.Error(s.searchHandler))

	writer := httptest.NewRecorder()
	formData := url.Values{}
	formData.Add("search", "test")
	request, err := http.NewRequest(http.MethodPost, "/search", strings.NewReader(formData.Encode()))
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
}
