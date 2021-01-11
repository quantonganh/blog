package mail

import (
	"github.com/quantonganh/blog/post"
)

type Mailer interface {
	SendConfirmationEmail(to, token string) error
	SendThankYouEmail(to string) error
	SendNewsletter(posts []*post.Post)
}
