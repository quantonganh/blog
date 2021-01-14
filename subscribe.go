package blog

type SubscribeService interface {
	FindByEmail(email string) (*Subscribe, error)
	Insert(s *Subscribe) error
	FindByToken(token string) (*Subscribe, error)
	FindByStatus(status string) ([]Subscribe, error)
	Subscribe(token string) error
	Unsubscribe(email string) error
}

type Subscribe struct {
	Email  string `bson:"email"`
	Token  string `bson:"token"`
	Status string `bson:"status"`
}

func NewSubscribe(email, token, status string) *Subscribe {
	return &Subscribe{
		Email:  email,
		Token:  token,
		Status: status,
	}
}
