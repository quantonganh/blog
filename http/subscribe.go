package http

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	gomongo "go.mongodb.org/mongo-driver/mongo"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/mongo"
)

const (
	confirmationMessage      = "A confirmation email has been sent to %s. Click the link in the email to confirm and activate your subscription. Check your spam folder if you don't see it within a couple of minutes."
	thankyouMessage          = "Thank you for subscribing to this blog."
	pendingMessage           = "Your subscription status is pending. Please click the confirmation link in your email."
	alreadySubscribedMessage = "You had been subscribed to this blog already."
)

func (s *Server) subscribeHandler(w http.ResponseWriter, r *http.Request) *AppError {
	email := r.FormValue("email")
	token := s.SMTPService.GenerateNewUUID()
	newSubscriber := blog.NewSubscribe(email, token, mongo.StatusPending)

	subscribe, err := s.SubscribeService.FindByEmail(email)
	if err != nil {
		if err == gomongo.ErrNoDocuments {
			if err := s.SMTPService.SendConfirmationEmail(email, token); err != nil {
				return &AppError{
					Error:   err,
					Message: "There is an error when sending confirmation email.",
					Code:    http.StatusInternalServerError,
				}
			}

			if err := s.SubscribeService.Insert(newSubscriber); err != nil {
				return &AppError{
					Error:   err,
					Message: "Failed to insert new subscriber.",
					Code:    http.StatusInternalServerError,
				}
			}

			if err := s.Renderer.RenderResponseMessage(w, fmt.Sprintf(confirmationMessage, newSubscriber.Email)); err != nil {
				return &AppError{
					Error:   err,
					Message: "There is an error when inserting new subscriber into database.",
					Code:    http.StatusInternalServerError,
				}
			}
		} else {
			return &AppError{
				Error: err,
				Code:  http.StatusNotFound,
			}
		}
	} else {
		switch subscribe.Status {
		case mongo.StatusPending:
			if err := s.Renderer.RenderResponseMessage(w, pendingMessage); err != nil {
				return &AppError{
					Error:   err,
					Message: "There is an error when rendering subscribe template.",
					Code:    http.StatusInternalServerError,
				}
			}
		case mongo.StatusSubscribed:
			if err := s.Renderer.RenderResponseMessage(w, alreadySubscribedMessage); err != nil {
				return &AppError{
					Error:   err,
					Message: "There is an error when rendering subscribe template.",
					Code:    http.StatusInternalServerError,
				}
			}
		default:
			if err := s.SubscribeService.Subscribe(subscribe.Token); err != nil {
				return &AppError{
					Error: err,
					Code:  http.StatusInternalServerError,
				}
			}

			if err := s.Renderer.RenderResponseMessage(w, resubscribedMessage); err != nil {
				return &AppError{
					Error:   err,
					Message: "There is an error when rendering subscribe template.",
					Code:    http.StatusInternalServerError,
				}
			}
		}
	}

	return nil
}

func (s *Server) confirmHandler(w http.ResponseWriter, r *http.Request) *AppError {
	token := r.URL.Query().Get("token")
	if len(token) == 0 {
		return &AppError{
			Error:   errors.New("token is not present"),
			Message: "Missing token",
			Code:    http.StatusNotFound,
		}
	}

	if err := s.SubscribeService.Subscribe(token); err != nil {
		return &AppError{
			Error:   err,
			Message: "failed to update subscribe status",
			Code:    http.StatusInternalServerError,
		}
	}

	subscribe, err := s.SubscribeService.FindByToken(token)
	if err != nil {
		return &AppError{
			Error:   err,
			Message: "Cannot find subscriber by token",
			Code:    http.StatusNotFound,
		}
	}

	if err := s.SMTPService.SendThankYouEmail(subscribe.Email); err != nil {
		return &AppError{
			Error:   err,
			Message: fmt.Sprintf("There is a problem when sending thank you email to %s", subscribe.Email),
			Code:    http.StatusInternalServerError,
		}
	}

	if err := s.Renderer.RenderResponseMessage(w, thankyouMessage); err != nil {
		return &AppError{
			Error:   err,
			Message: "failed to render subscribe template",
			Code:    http.StatusInternalServerError,
		}
	}

	return nil
}
