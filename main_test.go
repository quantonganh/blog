package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/quantonganh/blog/post"
)

func TestHandler(t *testing.T) {
	posts, err := post.GetAllPosts("posts")
	assert.NoError(t, err)
	assert.NotZero(t, len(posts))
	a := app{
		Blog: post.NewBlog(posts),
	}

	t.Run("homeHandler", func(t *testing.T) {
		a.testHomeHandler(t, posts)
	})

	t.Run("postHandler", a.testPostHandler)

	t.Run("tagHandler", a.testTagHandler)

	t.Run("searchHandler", a.testSearchHandler)

	t.Run("subscribeHandler", a.testSubscribeHandler)

	t.Run("confirmHandler", a.testConfirmHandler)

	t.Run("unsubscribeHandler", a.testUnsubscribeHandler)
}

func (a *app) testHomeHandler(t *testing.T, posts []*post.Post) {
	router := mux.NewRouter()
	router.HandleFunc("/", mwError(a.homeHandler(posts)))

	writer := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
}

func (a *app) testPostHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/{year:20[1-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}/{postName}", mwError(a.postHandler))

	writer := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/2019/09/19/about", nil)
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
}

func (a *app) testTagHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/tag/test", mwError(a.tagHandler))

	writer := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/tag/test", nil)
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
}

func (a *app) testSearchHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/search", mwError(a.searchHandler))

	writer := httptest.NewRecorder()
	formData := url.Values{}
	formData.Add("search", "test")
	request, err := http.NewRequest(http.MethodPost, "/search", strings.NewReader(formData.Encode()))
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
}
