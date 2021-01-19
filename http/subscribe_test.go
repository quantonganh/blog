package http

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/asdine/storm/v3"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/mock"
)

func (s *Server) testSubscribeHandler(t *testing.T) {
	email := "foo@gmail.com"
	token := uuid.NewV4().String()

	subscribe := &blog.Subscribe{}
	subscribeService := new(mock.SubscribeService)
	subscribeService.On("FindByEmail", email).Return(subscribe, storm.ErrNotFound)
	subscribeService.On("Insert", blog.NewSubscribe(email, token, blog.StatusPending)).Return(nil)

	smtpService := new(mock.SMTPService)
	smtpService.On("SendConfirmationEmail", email, token).Return(nil)
	smtpService.On("GenerateNewUUID").Return(token)

	s.SubscribeService = subscribeService
	s.SMTPService = smtpService

	router := mux.NewRouter()
	router.HandleFunc("/subscribe", s.Error(s.subscribeHandler)).Methods(http.MethodPost)

	form := url.Values{}
	form.Add("email", email)
	req, err := http.NewRequest(http.MethodPost, "/subscribe", strings.NewReader(form.Encode()))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, fmt.Sprintf(confirmationMessage, email), getResponseMessage(resp.Body))
}

func (s *Server) testConfirmHandler(t *testing.T) {
	email := "foo@gmail.com"
	token := uuid.NewV4().String()

	subscribe := blog.NewSubscribe(email, token, blog.StatusPending)
	subscribeService := new(mock.SubscribeService)
	subscribeService.On("Subscribe", token).Return(nil)
	subscribeService.On("FindByToken", token).Return(subscribe, nil)

	smtpService := new(mock.SMTPService)
	smtpService.On("SendThankYouEmail", email).Return(nil)

	s.SubscribeService = subscribeService
	s.SMTPService = smtpService

	router := mux.NewRouter()
	router.HandleFunc("/subscribe/confirm", s.Error(s.confirmHandler))

	form := url.Values{}
	form.Add("email", email)
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/subscribe/confirm?token=%s", token), nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, thankyouMessage, getResponseMessage(resp.Body))
}

func getResponseMessage(body io.ReadCloser) string {
	tokenizer := html.NewTokenizer(body)
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			if err := tokenizer.Err(); err == io.EOF {
				break
			}
		}

		token := tokenizer.Token()
		if token.Data == "p" {
			tokenType = tokenizer.Next()
			if tokenType == html.TextToken {
				return tokenizer.Token().Data
			}
		}
	}

	return ""
}
