package http

import (
	"fmt"
	"net/http"

	"github.com/asdine/storm/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/hlog"

	"github.com/quantonganh/blog"
)

const (
	confirmationMessage      = "A confirmation email has been sent to %s. Click the link in the email to confirm and activate your subscription. Check your spam folder if you don't see it within a couple of minutes."
	thankyouMessage          = "Thank you for subscribing to this blog."
	pendingMessage           = "Your subscription status is pending. Please click the confirmation link in your email."
	alreadySubscribedMessage = "You had been subscribed to this blog already."
)

func (s *Server) subscribeHandler(w http.ResponseWriter, r *http.Request) error {
	email := r.FormValue("email")
	token := s.SMTPService.GenerateNewUUID()
	newSubscriber := blog.NewSubscribe(email, token, blog.StatusPending)

	logger := hlog.FromRequest(r)
	subscribe, err := s.SubscribeService.FindByEmail(email)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			logger.Info().Msg("Sending confirmation email")
			if err := s.SMTPService.SendConfirmationEmail(email, token); err != nil {
				return err
			}

			logger.Info().Msgf("Saving new subscriber %+v into the database", newSubscriber)
			if err := s.SubscribeService.Insert(newSubscriber); err != nil {
				return err
			}

			logger.Info().Msg("Rendering the response message")
			if err := s.Renderer.RenderResponseMessage(w, fmt.Sprintf(confirmationMessage, newSubscriber.Email)); err != nil {
				return err
			}
		} else {
			return NewError(err, http.StatusNotFound, fmt.Sprintf("Cannot found email: %s", email))
		}
	} else {
		logger.Info().Msgf("Found subscriber %+v in the database", subscribe)
		switch subscribe.Status {
		case blog.StatusPending:
			if err := s.Renderer.RenderResponseMessage(w, pendingMessage); err != nil {
				return err
			}
		case blog.StatusSubscribed:
			if err := s.Renderer.RenderResponseMessage(w, alreadySubscribedMessage); err != nil {
				return err
			}
		default:
			if err := s.SMTPService.SendConfirmationEmail(email, token); err != nil {
				return err
			}

			logger.Info().Msgf("Updating status to %s", blog.StatusPending)
			if err := s.SubscribeService.Update(email, token); err != nil {
				return err
			}

			logger.Info().Msg("Rendering the response message")
			if err := s.Renderer.RenderResponseMessage(w, fmt.Sprintf(confirmationMessage, email)); err != nil {
				return err
			}
		}
	}

	return nil
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

	if err := s.Renderer.RenderResponseMessage(w, thankyouMessage); err != nil {
		return err
	}

	return nil
}
