package blog

// SubscribeService is the interface that wraps methods related to subscribe function
type SubscribeService interface {
	FindByEmail(email string) (*Subscribe, error)
	Insert(s *Subscribe) error
	UpdateStatus(email string) error
	FindByToken(token string) (*Subscribe, error)
	FindByStatus(status string) ([]Subscribe, error)
	Subscribe(token string) error
	Unsubscribe(email string) error
}

// Subscribe represents a subscriber
type Subscribe struct {
	ID     int    `storm:"id,increment"`
	Email  string `storm:"unique"`
	Token  string `storm:"index"`
	Status string `storm:"index"`
}

// Subscribe status
const (
	StatusPending      = "pending"
	StatusSubscribed   = "subscribed"
	StatusUnsubscribed = "unsubscribed"
)

// NewSubscribe returns new subscriber
func NewSubscribe(email, token, status string) *Subscribe {
	return &Subscribe{
		Email:  email,
		Token:  token,
		Status: status,
	}
}
