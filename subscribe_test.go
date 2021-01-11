package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/html"

	"github.com/quantonganh/blog/subscriber"
	mailMocks "github.com/quantonganh/blog/subscriber/mail/mocks"
	subscriberMocks "github.com/quantonganh/blog/subscriber/mocks"
)

func (a *app) testSubscribeHandler(t *testing.T) {
	email := "foo@gmail.com"
	token := uuid.NewV4().String()

	s := &subscriber.Subscriber{}
	ml := new(subscriberMocks.MailingList)
	ml.On("FindByEmail", email).Return(s, mongo.ErrNoDocuments)
	ml.On("Insert", subscriber.New(email, token)).Return(nil)

	mailer := new(mailMocks.Mailer)
	mailer.On("SendConfirmationEmail", email, token).Return(nil)

	a.MailingList = ml
	a.Mailer = mailer

	router := mux.NewRouter()
	router.HandleFunc("/subscribe", mwError(a.subscribeHandler(token))).Methods(http.MethodPost)

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

func (a *app) testConfirmHandler(t *testing.T) {
	email := "foo@gmail.com"
	token := uuid.NewV4().String()

	s := subscriber.New(email, token)
	ml := new(subscriberMocks.MailingList)
	ml.On("Subscribe", token).Return(nil)
	ml.On("FindByToken", token).Return(s, nil)

	mailer := new(mailMocks.Mailer)
	mailer.On("SendThankYouEmail", email).Return(nil)

	a.MailingList = ml
	a.Mailer = mailer

	router := mux.NewRouter()
	router.HandleFunc("/subscribe/confirm", mwError(a.confirmHandler))

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

		if err := tokenizer.Err(); err == io.EOF {
			break
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
