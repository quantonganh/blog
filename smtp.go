package blog

type SMTPService interface {
	SendConfirmationEmail(to, token string) error
	SendThankYouEmail(to string) error
	SendNewsletter(posts []*Post)
	GenerateNewUUID() string
	GetHMACSecret() string
}
