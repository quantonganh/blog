package blog

// SMTPService is the interface that wraps methods related to SMTP
type SMTPService interface {
	SendConfirmationEmail(to, token string) error
	SendThankYouEmail(to string) error
	SendNewsletter(posts []*Post)
	GenerateNewUUID() string
	GetHMACSecret() string
	Stop() error
}
