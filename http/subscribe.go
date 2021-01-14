package http

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	gomongo "go.mongodb.org/mongo-driver/mongo"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/http/html"
	"github.com/quantonganh/blog/http/mw"
	"github.com/quantonganh/blog/mongo"
)

const (
	confirmationMessage      = "A confirmation email has been sent to %s. Click the link in the email to confirm and activate your subscription. Check your spam folder if you don't see it within a couple of minutes."
	thankyouMessage          = "Thank you for subscribing to this blog."
	pendingMessage           = "Your subscription status is pending. Please click the confirmation link in your email."
	alreadySubscribedMessage = "You had been subscribed to this blog already."
)

func (s *Server) subscribeHandler(token string) mw.ErrHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		email := r.FormValue("email")
		newSubscriber := blog.NewSubscribe(email, token, mongo.StatusPending)

		tmpl := html.NewSubscribe(s.Templates)
		subscribe, err := s.SubscribeService.FindByEmail(email)
		if err != nil {
			if err == gomongo.ErrNoDocuments {
				if err := s.SMTPService.SendConfirmationEmail(email, token); err != nil {
					return err
				}

				if err := s.SubscribeService.Insert(newSubscriber); err != nil {
					return err
				}

				if err := tmpl.Render(w, fmt.Sprintf(confirmationMessage, newSubscriber.Email)); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			switch subscribe.Status {
			case mongo.StatusPending:
				if err := tmpl.Render(w, pendingMessage); err != nil {
					return err
				}
			case mongo.StatusSubscribed:
				if err := tmpl.Render(w, alreadySubscribedMessage); err != nil {
					return err
				}
			default:
				if err := s.SubscribeService.Subscribe(subscribe.Token); err != nil {
					return err
				}

				if err := tmpl.Render(w, resubscribedMessage); err != nil {
					return err
				}
			}
		}

		return nil
	}
}

func (s *Server) confirmHandler(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.Query().Get("token")
	if len(token) == 0 {
		return errors.New("token is not present")
	}

	if err := s.SubscribeService.Subscribe(token); err != nil {
		return err
	}

	subscribe, err := s.SubscribeService.FindByToken(token)
	if err != nil {
		return err
	}

	if err := s.SMTPService.SendThankYouEmail(subscribe.Email); err != nil {
		return err
	}

	return html.NewSubscribe(s.Templates).Render(w, thankyouMessage)
}
