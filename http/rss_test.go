package http

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/feeds"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRSSHandler(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/rss.xml", nil)
	assert.NoError(t, err)
	s.router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	var rss *feeds.RssFeedXml
	require.NoError(t, xml.NewDecoder(rr.Body).Decode(&rss))
	assert.Equal(t, 1, len(rss.Channel.Items))
	assert.Equal(t, "http://localhost/2019/09/19/test", rss.Channel.Items[0].Link)
}
