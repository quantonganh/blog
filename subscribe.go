package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/flosch/pongo2"
	"github.com/matcornic/hermes/v2"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	uuid "github.com/satori/go.uuid"
	gomail "gopkg.in/mail.v2"
)

const (
	subscribersFile = "subscribers.json"
)

type mailingList struct {
	mu          sync.RWMutex
	Subscribers []subscriber
}

type subscriber struct {
	Email   string
	Token   string
	Pending bool
}

func (ml *mailingList) Save(w io.Writer) error {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	subscribersJSON, err := json.MarshalIndent(ml.Subscribers, "", "    ")
	if err != nil {
		return errors.Errorf("failed to marshal: %v", err)
	}

	switch f := w.(type) {
	case *os.File:
		if err := f.Truncate(0); err != nil {
			return errors.Errorf("failed to truncate: %v", err)
		}
		_, err := f.Seek(0, 0)
		if err != nil {
			return errors.Errorf("failed to seek: %v", err)
		}
	default:
	}

	_, err = w.Write(subscribersJSON)
	if err != nil {
		return errors.Errorf("failed to write bytes to underlying data stream: %v", err)
	}

	return nil
}

func (ml *mailingList) Load(r io.Reader) error {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.Errorf("failed to read bytes into b: %v", err)
	}

	if len(b) > 0 {
		if err := json.Unmarshal(b, &ml.Subscribers); err != nil {
			return errors.Errorf("failed to unmarshal: %v", err)
		}
	}

	return nil
}

func (b *Blog) subscribeHandler(w http.ResponseWriter, r *http.Request) error {
	email := r.FormValue("email")
	token := uuid.NewV4().String()
	subscriber := subscriber{
		Email:   email,
		Token:   token,
		Pending: true,
	}

	f, err := os.OpenFile(subscribersFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errors.Errorf("failed to open file %s: %v", subscribersFile, err)
	}
	defer func() {
		_ = f.Close()
	}()

	fileInfo, err := os.Stat(subscribersFile)
	if err != nil {
		return errors.Errorf("failed to get file info of %s: %v", subscribersFile, err)
	}

	ml := &mailingList{}
	if fileInfo.Size() > 0 {
		if err := ml.Load(f); err != nil {
			return err
		}

		for _, s := range ml.Subscribers {
			if s.Email == subscriber.Email {
				return b.renderSubscribe(w, "You had been subscribed to this blog already.")
			}
		}
	}

	ml.Subscribers = append(ml.Subscribers, subscriber)
	if err := ml.Save(f); err != nil {
		return err
	}

	if err := b.sendConfirmationEmail(email, r.Host, token); err != nil {
		return err
	}

	message := fmt.Sprintf("A confirmation email has been sent to %s. Click the link in the email to confirm and activate your subscription. Check your spam folder if you don't see it within a couple of minutes.", subscriber.Email)
	err = b.renderSubscribe(w, message)
	if err != nil {
		return err
	}

	return nil
}

func (b *Blog) sendConfirmationEmail(to, host, token string) error {
	h := hermes.Hermes{
		Product: hermes.Product{
			Name: b.config.Newsletter.Product.Name,
			Link: fmt.Sprintf("http://%s", host),
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name: "",
			Intros: []string{
				fmt.Sprintf("Welcome to %s", b.config.Newsletter.Product.Name),
			},
			Actions: []hermes.Action{
				{
					Instructions: "",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Confirm your subscription",
						Link:  fmt.Sprintf("http://%s/subscribe/confirm?token=%s", host, token),
					},
				},
			},
		},
	}

	emailBody, err := h.GenerateHTML(email)
	if err != nil {
		return errors.Errorf("failed to generate HTML email: %v", err)
	}

	return b.sendEmail([]string{to}, "Confirm subscription", emailBody)
}

func (b *Blog) confirmHandler(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.Query().Get("token")
	if len(token) == 0 {
		return errors.New("token is not present")
	}

	f, err := os.OpenFile(subscribersFile, os.O_RDWR, 0644)
	if err != nil {
		return errors.Errorf("failed to open file %s: %v", subscribersFile, err)
	}
	defer func() {
		_ = f.Close()
	}()

	ml := &mailingList{}
	if err := ml.Load(f); err != nil {
		return err
	}

	for i, s := range ml.Subscribers {
		if s.Token == token {
			ml.Subscribers[i].Pending = false
			if err := ml.Save(f); err != nil {
				return err
			}

			if err := b.sendThankYouEmail(s.Email, r.Host); err != nil {
				return err
			}

			return b.renderSubscribe(w, "Thank you for subscribing to this blog.")
		}
	}

	return b.renderSubscribe(w, "Something went wrong.")
}

func (b *Blog) sendThankYouEmail(to, host string) error {
	h := hermes.Hermes{
		Product: hermes.Product{
			Name: b.config.Newsletter.Product.Name,
			Link: fmt.Sprintf("http://%s", host),
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name: "",
			Intros: []string{
				fmt.Sprintf("Thank you for subscribing to %s", b.config.Newsletter.Product.Name),
			},
			Actions: []hermes.Action{
				{
					Instructions: "You will receive updates to your inbox.",
				},
			},
		},
	}

	emailBody, err := h.GenerateHTML(email)
	if err != nil {
		return errors.Errorf("failed to generate HTML email: %v", err)
	}

	return b.sendEmail([]string{to}, "Thank you for subscribing", emailBody)
}

func (b *Blog) sendNewsletter(posts []*Post) {
	c := cron.New(
		cron.WithLogger(
			cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))
	_, _ = c.AddFunc(b.config.Cron.Spec, func() {
		fi, _ := os.Open(subscribersFile)
		defer func() {
			_ = fi.Close()
		}()

		ml := &mailingList{}
		if err := ml.Load(fi); err == nil {
			var (
				recipients []string
				buf        = new(bytes.Buffer)
			)
			for _, s := range ml.Subscribers {
				if !s.Pending {
					recipients = append(recipients, s.Email)

					pageURL, err := url.Parse(os.Getenv("PAGE_URL"))
					if err != nil {
						log.Fatal(err)
					}
					hash, err := computeHmac256(s.Email, b.config.HMAC.Secret)
					if err != nil {
						log.Fatal(err)
					}
					data := pongo2.Context{"posts": posts, "pageURL": pageURL, "email": s.Email, "hash": hash}
					if err := templates.ExecuteTemplate(buf, "newsletter", data); err != nil {
						log.Fatal(err)
					}
				}
			}

			if len(recipients) > 0 {
				if err := b.sendEmail(recipients, fmt.Sprintf("%s newsletter", b.config.Newsletter.Product.Name), buf.String()); err != nil {
					log.Fatal(err)
				}
			}
		}
	})

	c.Start()
}

func computeHmac256(message, secret string) (string, error) {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(message))
	if err != nil {
		return "", errors.Wrap(err, "hmac.Write")
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

func (b *Blog) sendEmail(to []string, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", b.config.SMTP.Username)
	addresses := make([]string, 0, len(to))
	for _, address := range to {
		addresses = append(addresses, m.FormatAddress(address, ""))
	}
	m.SetHeader("To", addresses...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := gomail.NewDialer(b.config.SMTP.Host, b.config.SMTP.Port, b.config.SMTP.Username, b.config.SMTP.Password)
	if err := d.DialAndSend(m); err != nil {
		return errors.Errorf("failed to send mail to %s: %v", fmt.Sprintf("%+v\n", to), err)
	}

	return nil
}

func (b *Blog) renderSubscribe(w http.ResponseWriter, message string) error {
	data := pongo2.Context{"message": message}
	if b.config != nil {
		data["navbarItems"] = b.config.Navbar.Items
	}

	if err := templates.ExecuteTemplate(w, "subscribe", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}