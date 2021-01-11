package main

import (
	"fmt"
	"net/http"

	"github.com/flosch/pongo2"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/quantonganh/blog/subscriber"
)

const (
	confirmationMessage = "A confirmation email has been sent to %s. Click the link in the email to confirm and activate your subscription. Check your spam folder if you don't see it within a couple of minutes."
	thankyouMessage     = "Thank you for subscribing to this blog."
)

func (a *app) subscribeHandler(token string) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		email := r.FormValue("email")
		newSubscriber := subscriber.New(email, token)

		s, err := a.MailingList.FindByEmail(email)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				if err := a.MailingList.Insert(newSubscriber); err != nil {
					return err
				}

				if err := a.Mailer.SendConfirmationEmail(email, token); err != nil {
					return err
				}

				if err := a.renderSubscribe(w, fmt.Sprintf(confirmationMessage, newSubscriber.Email)); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			switch s.Status {
			case subscriber.StatusPending:
				if err := a.renderSubscribe(w, "Your subscription status is pending. Please click the confirmation link in your email."); err != nil {
					return err
				}
			case subscriber.StatusSubscribed:
				if err := a.renderSubscribe(w, "You had been subscribed to this app already."); err != nil {
					return err
				}
			default:
				if err := a.MailingList.Subscribe(s.Token); err != nil {
					return err
				}

				if err := a.renderSubscribe(w, "You have been re-subscribed to this app."); err != nil {
					return err
				}
			}
		}

		return nil
	}
}

func (a *app) confirmHandler(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.Query().Get("token")
	if len(token) == 0 {
		return errors.New("token is not present")
	}

	if err := a.MailingList.Subscribe(token); err != nil {
		return err
	}

	s, err := a.MailingList.FindByToken(token)
	if err != nil {
		return err
	}

	if err := a.Mailer.SendThankYouEmail(s.Email); err != nil {
		return err
	}

	return a.renderSubscribe(w, thankyouMessage)
}

func (a *app) renderSubscribe(w http.ResponseWriter, message string) error {
	data := pongo2.Context{"message": message}
	if a.Config != nil {
		data["navbarItems"] = a.Config.Navbar.Items
	}

	if err := templates.ExecuteTemplate(w, "subscribe", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}
