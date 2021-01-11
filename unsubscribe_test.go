package main

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

	"github.com/quantonganh/blog/config"
	"github.com/quantonganh/blog/subscriber/hash"
	subscriberMocks "github.com/quantonganh/blog/subscriber/mocks"
)

func (a *app) testUnsubscribeHandler(t *testing.T) {
	viper.SetConfigType("yaml")
	var yamlConfig = []byte(`
newsletter:
  hmac:
    secret: da02e221bc331c9875c5e1299fa8d765
`)
	require.NoError(t, viper.ReadConfig(bytes.NewBuffer(yamlConfig)))
	var cfg *config.Config
	require.NoError(t, viper.Unmarshal(&cfg))
	a.Config = cfg

	email := "foo@gmail.com"
	hashValue, err := hash.ComputeHmac256(email, a.Config.Newsletter.HMAC.Secret)

	ml := new(subscriberMocks.MailingList)
	ml.On("Unsubscribe", email).Return(nil)

	a.MailingList = ml

	router := mux.NewRouter()
	router.HandleFunc("/unsubscribe", mwError(a.unsubscribeHandler))

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/unsubscribe?email=%s&hash=%s", email, hashValue), nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, unsubscribeMessage, getResponseMessage(resp.Body))
}
