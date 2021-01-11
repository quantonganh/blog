package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

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

type Subscriber struct {
	Email   string
	Token   string
	Pending bool
}

func Save(s []Subscriber, filename string) error {
	subscribersJSON, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return errors.Errorf("failed to marshal: %v", err)
	}

	if err := ioutil.WriteFile(filename, subscribersJSON, 0600); err != nil {
		return errors.Errorf("failed to write to file %s: %v", subscribersFile, err)
	}

	return nil
}

func Load(filename string) ([]Subscriber, error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, errors.Errorf("failed to open file %s: %v", subscribersFile, err)
	}
	defer func() {
		_ = jsonFile.Close()
	}()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, errors.Errorf("failed to read JSON file %s: %v", filename, err)
	}

	var s []Subscriber
	if err := json.Unmarshal(byteValue, &s); err != nil {
		return nil, errors.Errorf("failed to unmarshal: %v", err)
	}

	return s, nil
}

func (b *Blog) subscribeHandler(w http.ResponseWriter, r *http.Request) error {
	email := r.FormValue("email")
	token := uuid.NewV4().String()
	subscriber := Subscriber{
		Email:   email,
		Token:   token,
		Pending: true,
	}

	var (
		subscribers []Subscriber
		err         error
	)
	if _, err = os.Stat(subscribersFile); os.IsNotExist(err) {
	} else if err == nil {
		subscribers, err = Load(subscribersFile)
		if err != nil {
			return errors.Errorf("failed to load file %s: %v", subscribersFile, err)
		}

		for _, s := range subscribers {
			if s.Email == subscriber.Email {
				message := fmt.Sprintf("You had been subscribed to this blog already.")
				return b.renderSubscribe(w, message)
			}
		}
	} else {
		return errors.Wrap(err, "os.Stat")
	}

	subscribers = append(subscribers, subscriber)
	if err := Save(subscribers, subscribersFile); err != nil {
		return errors.Errorf("failed to save file %s: %v", subscribersFile, err)
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
			Name: "quantonganh.com blog",
			Link: "https://quantonganh.com",
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name: "",
			Intros: []string{
				"Welcome to quantonganh.com blog",
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

	return b.sendEmail([]string{to}, "quantonganh.com blog newsletter - confirm subscription", emailBody)
}

func (b *Blog) sendThankYouEmail(to string) error {
	h := hermes.Hermes{
		Product: hermes.Product{
			Name: "quantonganh.com blog",
			Link: "https://quantonganh.com",
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name: "",
			Intros: []string{
				"Thank you for subscribing to quantonganh.com blog",
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
	c.AddFunc("@every 0h0m10s", func() {
		subscribers, err := Load(subscribersFile)
		if err == nil {
			var (
				recipients []string
				buf        = new(bytes.Buffer)
			)
			for _, subscriber := range subscribers {
				if subscriber.Pending == false {
					recipients = append(recipients, subscriber.Email)

					pageURL, err := url.Parse(os.Getenv("PAGE_URL"))
					if err != nil {
						log.Fatal(err)
					}
					data := pongo2.Context{"posts": posts, "pageURL": pageURL, "email": subscriber.Email, "hash": computeHmac256(subscriber.Email, b.config.HMAC.Secret)}
					if err := templates.ExecuteTemplate(buf, "newsletter", data); err != nil {
						log.Fatal(err)
					}
				}
			}

			if len(recipients) > 0 {
				if err := b.sendEmail(recipients, "quantonganh.com blog newsletter", buf.String()); err != nil {
					log.Fatal(err)
				}
			}
		}
	})

	c.Start()
}

func computeHmac256(message, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
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

func (b *Blog) confirmHandler(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.Query().Get("token")
	if len(token) == 0 {
		return errors.New("token is not present")
	}

	subscribers, err := Load(subscribersFile)
	if err != nil {
		return errors.Errorf("failed to load file %subscribers: %v", subscribersFile, err)
	}

	for i, subscriber := range subscribers {
		if subscriber.Token == token {
			subscribers[i].Pending = false
			if err := Save(subscribers, subscribersFile); err != nil {
				return errors.Errorf("failed to save file %subscribers: %v", subscribersFile, err)
			}

			fmt.Printf("subscriber.Email: %subscribers\n", subscriber.Email)
			if err := b.sendThankYouEmail(subscriber.Email); err != nil {
				return err
			}

			return b.renderSubscribe(w, fmt.Sprintf("Thank you for subscribing to this blog."))
		}
	}

	return b.renderSubscribe(w, fmt.Sprintf("Something went wrong."))
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
