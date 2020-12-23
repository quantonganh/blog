package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	setUp()

	posts, err := getAllPosts("posts/2019/10/test.md")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(posts))

	t.Run("getAllPosts", func(t *testing.T) {
		testGetAllPosts(t, posts)
	})

	t.Run("homeHandler", func(t *testing.T) {
		testHomeHandler(t, posts)
	})

	t.Run("postHandler", func(t *testing.T) {
		testPostHandler(t, posts)
	})

	t.Run("tagHandler", func(t *testing.T) {
		testTagHandler(t, posts)
	})

	t.Cleanup(func() {
		_ = os.Remove("posts/2019/10/test.md")
	})
}

func setUp() {
	s := `---
title: title
date: Wed Oct  2 05:22:37 +07 2019
description: description
tags:
  - test
---
content`
	if err := os.MkdirAll("posts/2019/10", 0700); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("posts/2019/10/test.md", []byte(s), 0644); err != nil {
		log.Fatal(err)
	}
}

func testGetAllPosts(t *testing.T, posts []*Post) {
	p := posts[0]
	require.NotNil(t, p)
	assert.Equal(t, "title", p.Title)
	assert.Equal(t, "2019-10-02", toISODate(p.Date))
	assert.Equal(t, "description", p.Description)
	assert.Equal(t, 1, len(p.Tags))
	assert.Equal(t, "test", p.Tags[0])
	assert.Equal(t, template.HTML("<p>content</p>\n"), p.Content)
}

func testHomeHandler(t *testing.T, posts []*Post) {
	router := mux.NewRouter()
	router.HandleFunc("/", homeHandler(posts))

	writer := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	if writer.Code != http.StatusOK {
		t.Errorf("Response code is %v", writer.Code)
	}
}

func testPostHandler(t *testing.T, posts []*Post) {
	router := mux.NewRouter()
	router.HandleFunc("/{year:20[1-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}/{postName}", postHandler(posts))

	writer := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/2019/10/02/test", nil)
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	if writer.Code != http.StatusOK {
		t.Errorf("Response code is %v", writer.Code)
	}
}

func testTagHandler(t *testing.T, posts []*Post) {
	router := mux.NewRouter()
	router.HandleFunc("/tag/test", tagHandler(posts))

	writer := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/tag/test", nil)
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	if writer.Code != http.StatusOK {
		t.Errorf("Response code is %v", writer.Code)
	}
}

