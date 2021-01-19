package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantonganh/blog/mock"
	"github.com/quantonganh/blog/pkg/hash"
)

func (s *Server) testUnsubscribeHandler(t *testing.T, secret string) {
	email := "foo@gmail.com"
	hashValue, err := hash.ComputeHmac256(email, secret)
	require.NoError(t, err)

	subscribeService := new(mock.SubscribeService)
	subscribeService.On("Unsubscribe", email).Return(nil)

	s.SubscribeService = subscribeService

	smtpService := new(mock.SMTPService)
	smtpService.On("GetHMACSecret").Return(secret)
	s.SMTPService = smtpService

	router := mux.NewRouter()
	router.HandleFunc("/unsubscribe", s.Error(s.unsubscribeHandler))

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/unsubscribe?email=%s&hash=%s", email, hashValue), nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, unsubscribeMessage, getResponseMessage(resp.Body))
}
