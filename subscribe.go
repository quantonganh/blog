package blog

type SubscribeService interface {
	FindByEmail(email string) (*Subscribe, error)
	Insert(s *Subscribe) error
	UpdateStatus(email string) error
	FindByToken(token string) (*Subscribe, error)
	FindByStatus(status string) ([]Subscribe, error)
	Subscribe(token string) error
	Unsubscribe(email string) error
}

type Subscribe struct {
	ID     int    `storm:"id,increment"`
	Email  string `storm:"unique"`
	Token  string `storm:"index"`
	Status string `storm:"index"`
}

const (
	StatusPending      = "pending"
	StatusSubscribed   = "subscribed"
	StatusUnsubscribed = "unsubscribed"
)

func NewSubscribe(email, token, status string) *Subscribe {
	return &Subscribe{
		Email:  email,
		Token:  token,
		Status: status,
	}
}
