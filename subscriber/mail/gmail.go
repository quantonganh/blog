package mail

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os"

	"github.com/flosch/pongo2"
	"github.com/matcornic/hermes/v2"
	"github.com/pkg/errors"
	"github.com/quantonganh/blog/config"
	"github.com/quantonganh/blog/subscriber/hash"
	"github.com/robfig/cron/v3"
	"gopkg.in/gomail.v2"

	"github.com/quantonganh/blog/post"
	"github.com/quantonganh/blog/subscriber"
)

type gmail struct {
	pageURL *url.URL
	*config.Config
	*template.Template
	subscriber.MailingList
}

func NewGmail(pageURL *url.URL, config *config.Config, template *template.Template, ml subscriber.MailingList) *gmail {
	return &gmail{
		pageURL:     pageURL,
		Config:      config,
		Template:    template,
		MailingList: ml,
	}
}

func (g *gmail) SendConfirmationEmail(to, token string) error {
	h := hermes.Hermes{
		Product: hermes.Product{
			Name: g.Config.Newsletter.Product.Name,
			Link: g.pageURL.String(),
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name: "",
			Intros: []string{
				fmt.Sprintf("Welcome to %s", g.Config.Newsletter.Product.Name),
			},
			Actions: []hermes.Action{
				{
					Instructions: "",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Confirm your subscription",
						Link:  fmt.Sprintf("%s/subscribe/confirm?token=%s", g.pageURL.String(), token),
					},
				},
			},
		},
	}

	emailBody, err := h.GenerateHTML(email)
	if err != nil {
		return errors.Errorf("failed to generate HTML email: %v", err)
	}

	return g.sendEmail([]string{to}, "Confirm subscription", emailBody)
}

func (g *gmail) SendThankYouEmail(to string) error {
	h := hermes.Hermes{
		Product: hermes.Product{
			Name: g.Config.Newsletter.Product.Name,
			Link: g.pageURL.String(),
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name: "",
			Intros: []string{
				fmt.Sprintf("Thank you for subscribing to %s", g.Config.Newsletter.Product.Name),
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

	return g.sendEmail([]string{to}, "Thank you for subscribing", emailBody)
}

func (g *gmail) SendNewsletter(posts []*post.Post) {
	c := cron.New(
		cron.WithLogger(
			cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))
	_, _ = c.AddFunc(g.Config.Newsletter.Cron.Spec, func() {

		var (
			recipients []string
			buf        = new(bytes.Buffer)
		)
		subscribers, err := g.MailingList.FindByStatus(subscriber.StatusSubscribed)
		if err != nil {
			log.Fatal(err)
		}

		for _, s := range subscribers {
			recipients = append(recipients, s.Email)

			hash, err := hash.ComputeHmac256(s.Email, g.Config.Newsletter.HMAC.Secret)
			if err != nil {
				log.Fatal(err)
			}
			data := pongo2.Context{"posts": posts, "pageURL": g.pageURL, "email": s.Email, "hash": hash}
			if err := g.Template.ExecuteTemplate(buf, "newsletter", data); err != nil {
				log.Fatal(err)
			}
		}

		if len(recipients) > 0 {
			if err := g.sendEmail(recipients, fmt.Sprintf("%s newsletter", g.Config.Newsletter.Product.Name), buf.String()); err != nil {
				log.Fatal(err)
			}
		}
	})

	c.Start()
}

func (g *gmail) sendEmail(to []string, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", g.Config.SMTP.Username)
	addresses := make([]string, 0, len(to))
	for _, address := range to {
		addresses = append(addresses, m.FormatAddress(address, ""))
	}
	m.SetHeader("To", addresses...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := gomail.NewDialer(g.Config.SMTP.Host, g.Config.SMTP.Port, g.Config.SMTP.Username, g.Config.SMTP.Password)
	if err := d.DialAndSend(m); err != nil {
		return errors.Errorf("failed to send mail to %s: %v", fmt.Sprintf("%+v\n", to), err)
	}

	return nil
}
