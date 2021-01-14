package http

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/http/mw"
	"github.com/quantonganh/blog/mock"
	"github.com/quantonganh/blog/pkg/hash"
)

func (s *Server) testUnsubscribeHandler(t *testing.T) {
	viper.SetConfigType("yaml")
	var yamlConfig = []byte(`
newsletter:
  hmac:
    secret: da02e221bc331c9875c5e1299fa8d765
`)
	require.NoError(t, viper.ReadConfig(bytes.NewBuffer(yamlConfig)))
	var cfg *blog.Config
	require.NoError(t, viper.Unmarshal(&cfg))

	email := "foo@gmail.com"
	hashValue, err := hash.ComputeHmac256(email, cfg.Newsletter.HMAC.Secret)
	require.NoError(t, err)

	subscribeService := new(mock.SubscribeService)
	subscribeService.On("Unsubscribe", email).Return(nil)

	s.SubscribeService = subscribeService

	router := mux.NewRouter()
	router.HandleFunc("/unsubscribe", mw.Error(s.unsubscribeHandler(cfg.Newsletter.HMAC.Secret)))

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/unsubscribe?email=%s&hash=%s", email, hashValue), nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, unsubscribeMessage, getResponseMessage(resp.Body))
}
