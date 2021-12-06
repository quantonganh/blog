package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTagsHandler(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/tags", nil)
	assert.NoError(t, err)
	s.router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "/tag/test", getLinkByText(t, rr.Body, "test (1)"))
}
