package http

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/asdine/storm/v3"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	nethtml "golang.org/x/net/html"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/mock"
	"github.com/quantonganh/blog/ondisk"
	"github.com/quantonganh/blog/pkg/hash"
)

const indexPath = "test.bleve"

var (
	cfg  *blog.Config
	post *blog.Post
	s    *Server
)

func TestMain(m *testing.M) {
	viper.SetConfigType("yaml")
	var yamlConfig = []byte(`
templates:
  dir: html/templates

newsletter:
  hmac:
    secret: da02e221bc331c9875c5e1299fa8d765
`)
	if err := viper.ReadConfig(bytes.NewBuffer(yamlConfig)); err != nil {
		log.Fatal(err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}

	r := strings.NewReader(`---
title: Test
date: Thu Sep 19 21:48:39 +07 2019
description: Just a test
tags:
  - test
---
Test.`)
	var err error
	post, err = ondisk.ParseMarkdown(context.Background(), r)
	if err != nil {
		log.Fatal(err)
	}

	s, err = NewServer(cfg, []*blog.Post{post}, indexPath)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestParseMarkdown(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Test", post.Title)
	assert.Equal(t, "2019-09-19", blog.ToISODate(post.Date))
	assert.Equal(t, "Just a test", post.Description)
	assert.Equal(t, "test", post.Tags[0])
	assert.Equal(t, template.HTML("<p>Test.</p>\n"), post.Content)
}

func TestHomeHandler(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	s.router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "/2019/09/19/test", getLinkByText(t, rr.Body, "Test"))
}

func TestPostHandler(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/2019/09/19/test", nil)
	assert.NoError(t, err)
	s.router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "/tag/test", getLinkByText(t, rr.Body, "test"))
}

func TestTagHandler(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/tag/test", nil)
	assert.NoError(t, err)
	s.router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "/2019/09/19/test", getLinkByText(t, rr.Body, "Test"))
}

func TestSearchHandler(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	formData := url.Values{}
	formData.Add("q", "test")
	request, err := http.NewRequest(http.MethodPost, "/search", strings.NewReader(formData.Encode()))
	assert.NoError(t, err)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s.router.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "/2019/09/19/test", getLinkByText(t, rr.Body, "Test"))

	t.Cleanup(func() {
		_ = os.RemoveAll(indexPath)
	})
}

func getLinkByText(t *testing.T, body *bytes.Buffer, text string) string {
	doc, err := goquery.NewDocumentFromReader(body)
	require.NoError(t, err)

	var link string
	doc.Find("article a").Each(func(_ int, s *goquery.Selection) {
		if s.Text() == text {
			link, _ = s.Attr("href")
		}
	})

	return link
}

func TestSubscribeHandler(t *testing.T) {
	t.Parallel()

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

	form := url.Values{}
	form.Add("email", email)
	req, err := http.NewRequest(http.MethodPost, "/subscribe", strings.NewReader(form.Encode()))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, fmt.Sprintf(confirmationMessage, email), getResponseMessage(resp.Body))
}

func TestConfirmHandler(t *testing.T) {
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

	form := url.Values{}
	form.Add("email", email)
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/subscribe/confirm?token=%s", token), nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, thankyouMessage, getResponseMessage(resp.Body))
}

func getResponseMessage(body io.ReadCloser) string {
	tokenizer := nethtml.NewTokenizer(body)
	for {
		tokenType := tokenizer.Next()
		if tokenType == nethtml.ErrorToken {
			if err := tokenizer.Err(); err == io.EOF {
				break
			}
		}

		token := tokenizer.Token()
		if token.Data == "p" {
			tokenType = tokenizer.Next()
			if tokenType == nethtml.TextToken {
				return tokenizer.Token().Data
			}
		}
	}

	return ""
}

func TestUnsubscribeHandler(t *testing.T) {
	email := "foo@gmail.com"
	secret := cfg.Newsletter.HMAC.Secret
	hashValue, err := hash.ComputeHmac256(email, secret)
	require.NoError(t, err)

	subscribeService := new(mock.SubscribeService)
	subscribeService.On("Unsubscribe", email).Return(nil)

	s.SubscribeService = subscribeService

	smtpService := new(mock.SMTPService)
	smtpService.On("GetHMACSecret").Return(secret)
	s.SMTPService = smtpService

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/unsubscribe?email=%s&hash=%s", email, hashValue), nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, unsubscribeMessage, getResponseMessage(resp.Body))
}
