package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostByYearHandler(t *testing.T) {
	t.Run("by year", func(t *testing.T) {
		testPostHandler(t, "/2019")
	})
}

func TestPostByMonthHandler(t *testing.T) {
	t.Run("by month", func(t *testing.T) {
		testPostHandler(t, "/2019/09")
	})
}

func TestPostByDateHandler(t *testing.T) {
	t.Run("by date", func(t *testing.T) {
		testPostHandler(t, "/2019/09/19")
	})
}

func testPostHandler(t *testing.T, url string) {
	t.Parallel()

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)
	s.router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "/2019/09/19/test.md", getLinkByText(t, rr.Body, "Test"))
}
