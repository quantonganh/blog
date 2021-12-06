package http

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantonganh/blog"
)

func TestSitemapHandler(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	assert.NoError(t, err)
	s.router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	var sitemap *blog.Sitemap
	require.NoError(t, xml.NewDecoder(rr.Body).Decode(&sitemap))
	assert.Equal(t, 2, len(sitemap.URLs))
	assert.Equal(t, "http://localhost", sitemap.URLs[0].Loc)
	assert.Equal(t, "http://localhost/2019/09/19/test", sitemap.URLs[1].Loc)
}
