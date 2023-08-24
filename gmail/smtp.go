package gmail

import (
	"fmt"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/matcornic/hermes/v2"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/gomail.v2"

	"github.com/quantonganh/blog"
)

type smtpService struct {
	ServerURL string
	*blog.Config
	blog.SubscribeService
	blog.Renderer
	*cron.Cron
}

// NewSMTPService returns new SMTP service
func NewSMTPService(config *blog.Config, serverURL string, subscribeService blog.SubscribeService, renderer blog.Renderer) blog.SMTPService {
	return &smtpService{
		Config:           config,
		ServerURL:        serverURL,
		SubscribeService: subscribeService,
		Renderer:         renderer,
		Cron:             cron.New(cron.WithLogger(cron.DefaultLogger)),
	}
}

// SendConfirmationEmail sends a confirmation email
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

// SendThankYouEmail sends a "thank you" email
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

// SendNewsletter sends newsletter
func (smtp *smtpService) SendNewsletter(latestPosts []*blog.Post) {
	_, err := smtp.Cron.AddFunc(smtp.Config.Newsletter.Cron.Spec, func() {

		subscribers, err := smtp.SubscribeService.FindByStatus(blog.StatusSubscribed)
		if err != nil {
			sentry.CaptureException(err)
		}

		for _, s := range subscribers {
			buf, err := smtp.Renderer.RenderNewsletter(latestPosts, smtp.ServerURL, s.Email)
			if err != nil {
				sentry.CaptureException(err)
			}

			if err := smtp.sendEmail(s.Email, fmt.Sprintf("%s newsletter", smtp.Config.Newsletter.Product.Name), buf.String()); err != nil {
				sentry.CaptureException(err)
			}
		}
	})
	if err != nil {
		sentry.CaptureException(err)
	}

	smtp.Cron.Start()
}

// Stop stops SMTP service
func (smtp *smtpService) Stop() error {
	ctx := smtp.Cron.Stop()
	log.Println("Shutting down cron...")
	select {
	case <-time.After(10 * time.Second):
		return errors.New("cron forced to shutdown")
	case <-ctx.Done():
		log.Println("Cron exiting...")
		return ctx.Err()
	}
}

func (smtp *smtpService) sendEmail(to string, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", smtp.Config.Newsletter.From)
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

// GetHMACSecret gets HMAC secret from config
func (smtp *smtpService) GetHMACSecret() string {
	return smtp.Config.Newsletter.HMAC.Secret
}
