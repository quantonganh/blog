package gmail

import (
	"fmt"
	"log"
	"os"

	"github.com/matcornic/hermes/v2"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/gomail.v2"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/mongo"
)

type smtpService struct {
	ServerURL string
	*blog.Config
	blog.SubscribeService
	blog.Renderer
}

func NewSMTPService(config *blog.Config, serverURL string, subscribeService blog.SubscribeService, renderer blog.Renderer) *smtpService {
	return &smtpService{
		Config:           config,
		ServerURL:        serverURL,
		SubscribeService: subscribeService,
		Renderer:         renderer,
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

	return smtp.sendEmail(to, "Confirm subscription", emailBody)
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

	return smtp.sendEmail(to, "Thank you for subscribing", emailBody)
}

func (smtp *smtpService) SendNewsletter(latestPosts []*blog.Post) {
	c := cron.New(
		cron.WithLogger(
			cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))
	_, _ = c.AddFunc(smtp.Config.Newsletter.Cron.Spec, func() {

		subscribers, err := smtp.SubscribeService.FindByStatus(mongo.StatusSubscribed)
		if err != nil {
			log.Fatal(err)
		}

		for _, s := range subscribers {
			buf, err := smtp.Renderer.RenderNewsletter(latestPosts, smtp.ServerURL, s.Email)
			if err != nil {
				log.Fatal(err)
			}

			if err := smtp.sendEmail(s.Email, fmt.Sprintf("%s newsletter", smtp.Config.Newsletter.Product.Name), buf.String()); err != nil {
				log.Fatal(err)
			}
		}
	})

	c.Start()
}

func (smtp *smtpService) sendEmail(to string, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", smtp.Config.SMTP.Username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := gomail.NewDialer(smtp.Config.SMTP.Host, smtp.Config.SMTP.Port, smtp.Config.SMTP.Username, smtp.Config.SMTP.Password)
	if err := d.DialAndSend(m); err != nil {
		return errors.Errorf("failed to send mail to %s: %v", fmt.Sprintf("%+v\n", to), err)
	}

	return nil
}

func (smtp *smtpService) GenerateNewUUID() string {
	return uuid.NewV4().String()
}

func (smtp *smtpService) GetHMACSecret() string {
	return smtp.Config.Newsletter.HMAC.Secret
}
