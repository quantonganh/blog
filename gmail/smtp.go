package gmail

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/flosch/pongo2"
	"github.com/matcornic/hermes/v2"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"gopkg.in/gomail.v2"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/mongo"
	"github.com/quantonganh/blog/pkg/hash"
)

type smtpService struct {
	ServerURL string
	*blog.Config
	*template.Template
	blog.SubscribeService
}

func NewSMTPService(config *blog.Config, serverURL string, template *template.Template, subscribeService blog.SubscribeService) *smtpService {
	return &smtpService{
		Config:           config,
		ServerURL:        serverURL,
		Template:         template,
		SubscribeService: subscribeService,
	}
}

func (smtp *smtpService) SendConfirmationEmail(to, token string) error {
	h := hermes.Hermes{
		Product: hermes.Product{
			Name: smtp.Config.Newsletter.Product.Name,
			Link: smtp.ServerURL,
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name: "",
			Intros: []string{
				fmt.Sprintf("Welcome to %s", smtp.Config.Newsletter.Product.Name),
			},
			Actions: []hermes.Action{
				{
					Instructions: "",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Confirm your subscription",
						Link:  fmt.Sprintf("%s/subscribe/confirm?token=%s", smtp.ServerURL, token),
					},
				},
			},
		},
	}

	emailBody, err := h.GenerateHTML(email)
	if err != nil {
		return errors.Errorf("failed to generate HTML email: %v", err)
	}

	return smtp.sendEmail([]string{to}, "Confirm subscription", emailBody)
}

func (smtp *smtpService) SendThankYouEmail(to string) error {
	h := hermes.Hermes{
		Product: hermes.Product{
			Name: smtp.Config.Newsletter.Product.Name,
			Link: smtp.ServerURL,
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name: "",
			Intros: []string{
				fmt.Sprintf("Thank you for subscribing to %s", smtp.Config.Newsletter.Product.Name),
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

	return smtp.sendEmail([]string{to}, "Thank you for subscribing", emailBody)
}

func (smtp *smtpService) SendNewsletter(posts []*blog.Post) {
	c := cron.New(
		cron.WithLogger(
			cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))
	_, _ = c.AddFunc(smtp.Config.Newsletter.Cron.Spec, func() {

		var (
			recipients []string
			buf        = new(bytes.Buffer)
		)
		subscribers, err := smtp.SubscribeService.FindByStatus(mongo.StatusSubscribed)
		if err != nil {
			log.Fatal(err)
		}

		for _, s := range subscribers {
			recipients = append(recipients, s.Email)

			hash, err := hash.ComputeHmac256(s.Email, smtp.Config.Newsletter.HMAC.Secret)
			if err != nil {
				log.Fatal(err)
			}
			data := pongo2.Context{"posts": posts, "pageURL": smtp.ServerURL, "email": s.Email, "hash": hash}
			if err := smtp.Template.ExecuteTemplate(buf, "newsletter", data); err != nil {
				log.Fatal(err)
			}
		}

		if len(recipients) > 0 {
			if err := smtp.sendEmail(recipients, fmt.Sprintf("%s newsletter", smtp.Config.Newsletter.Product.Name), buf.String()); err != nil {
				log.Fatal(err)
			}
		}
	})

	c.Start()
}

func (smtp *smtpService) sendEmail(to []string, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", smtp.Config.SMTP.Username)
	addresses := make([]string, 0, len(to))
	for _, address := range to {
		addresses = append(addresses, m.FormatAddress(address, ""))
	}
	m.SetHeader("To", addresses...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := gomail.NewDialer(smtp.Config.SMTP.Host, smtp.Config.SMTP.Port, smtp.Config.SMTP.Username, smtp.Config.SMTP.Password)
	if err := d.DialAndSend(m); err != nil {
		return errors.Errorf("failed to send mail to %s: %v", fmt.Sprintf("%+v\n", to), err)
	}

	return nil
}
