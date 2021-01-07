package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	posts, err := getAllPosts("posts")
	assert.NoError(t, err)
	assert.NotZero(t, len(posts))
	b := Blog{
		posts: posts,
	}

	t.Run("homeHandler", b.testHomeHandler)

	t.Run("postHandler", b.testPostHandler)

	t.Run("tagHandler", b.testTagHandler)

	t.Run("searchHandler", b.testSearchHandler)
}

func (b *Blog) testHomeHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/", mwError(b.homeHandler))

	writer := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
}

func (b *Blog) testPostHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/{year:20[1-9][0-9]}/{month:0[1-9]|1[012]}/{day:0[1-9]|[12][0-9]|3[01]}/{postName}", mwError(b.postHandler))

	writer := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/2019/09/19/about", nil)
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
}

func (b *Blog) testTagHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/tag/test", mwError(b.tagHandler))

	writer := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/tag/test", nil)
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
}

func (b *Blog) testSearchHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/search", mwError(b.searchHandler))

	writer := httptest.NewRecorder()
	formData := url.Values{}
	formData.Add("search", "test")
	request, err := http.NewRequest(http.MethodPost, "/search", strings.NewReader(formData.Encode()))
	assert.NoError(t, err)
	router.ServeHTTP(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
}
