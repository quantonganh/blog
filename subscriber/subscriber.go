package subscriber

type MailingList interface {
	FindByEmail(email string) (*Subscriber, error)
	Insert(s *Subscriber) error
	FindByToken(token string) (*Subscriber, error)
	FindByStatus(status string) ([]Subscriber, error)
	Subscribe(token string) error
	Unsubscribe(email string) error
}

type Subscriber struct {
	Email  string `bson:"email"`
	Token  string `bson:"token"`
	Status string `bson:"status"`
}

func New(email, token string) *Subscriber {
	return &Subscriber{
		Email:  email,
		Token:  token,
		Status: StatusPending,
	}
}
