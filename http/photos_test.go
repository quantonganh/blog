package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhotosHandler(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/photos", nil)
	assert.NoError(t, err)
	s.router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "/2019/09/19/test", getLinkByImg(t, rr.Body, "/path/to/photo.jpg"))
}

func getLinkByImg(t *testing.T, body *bytes.Buffer, imgSrc string) string {
	doc, err := goquery.NewDocumentFromReader(body)
	require.NoError(t, err)

	var link string
	doc.Find("article .container-fluid .grid .grid-item a").Each(func(_ int, s *goquery.Selection) {
		src, _ := s.Find("img").Attr("src")
		if src == imgSrc {
			link, _ = s.Attr("href")
		}
	})

	return link
}
