package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	formData := url.Values{}
	formData.Add("q", "test")
	r, err := http.NewRequest(http.MethodPost, "/search", strings.NewReader(formData.Encode()))
	assert.NoError(t, err)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	testCases := []struct {
		name string
		err  error
	}{
		{
			name: "client error",
			err:  NewError(errors.New("Bad request: invalid search query"), http.StatusBadRequest, "invalid URL escape"),
		},
		{
			name: "server error",
			err:  errors.New("internal server error"),
		},
	}
	for _, tc := range testCases {
		rr := httptest.NewRecorder()

		t.Run(tc.name, func(t *testing.T) {
			fn := func(w http.ResponseWriter, r *http.Request) error {
				return tc.err
			}
			s.Error(fn).ServeHTTP(rr, r)

			rs := rr.Result()

			_, ok := tc.err.(ClientError)
			if ok {
				assert.Equal(t, http.StatusBadRequest, rr.Code)

				contentTypeHeaders := rs.Header["Content-Type"]
				assert.Equal(t, 1, len(contentTypeHeaders))
				assert.Equal(t, "application/json; charset=utf-8", contentTypeHeaders[0])

				var respErr Error
				require.NoError(t, json.NewDecoder(rs.Body).Decode(&respErr))
				assert.Equal(t, "invalid URL escape", respErr.Message)
			} else {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				msg, err := getResponseMessage(rs.Body)
				require.NoError(t, err)
				assert.Equal(t, errOops, strings.TrimSpace(msg))
			}
		})
	}
}
