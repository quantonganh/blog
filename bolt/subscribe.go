package bolt

import (
	"github.com/go-errors/errors"

	"github.com/quantonganh/blog"
)

type subscribeService struct {
	db *DB
}

func NewSubscribeService(db *DB) blog.SubscribeService {
	return &subscribeService{
		db: db,
	}
}

// FindByEmail finds a subscriber by email
func (ss *subscribeService) FindByEmail(email string) (*blog.Subscribe, error) {
	var s blog.Subscribe
	if err := ss.db.stormDB.One("Email", email, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

// Insert inserts new subscriber into stormDB
func (ss *subscribeService) Insert(s *blog.Subscribe) error {
	if err := ss.db.stormDB.Save(s); err != nil {
		return errors.Errorf("failed to save: %v", err)
	}

	return nil
}

// Update updates subscribe status and new token
func (ss *subscribeService) Update(email, token string) error {
	s, err := ss.FindByEmail(email)
	if err != nil {
		return err
	}

	s.Status = blog.StatusPending
	s.Token = token
	if err := ss.db.stormDB.Save(s); err != nil {
		return errors.Errorf("failed to save: %v", err)
	}

	return nil
}

// FindByToken finds subscriber by token
func (ss *subscribeService) FindByToken(token string) (*blog.Subscribe, error) {
	var s blog.Subscribe
	if err := ss.db.stormDB.One("Token", token, &s); err != nil {
		return nil, errors.Errorf("failed to find by token: %v", err)
	}

	return &s, nil
}

// FindByStatus finds subscriber by status
func (ss *subscribeService) FindByStatus(status string) ([]blog.Subscribe, error) {
	var subscribes []blog.Subscribe
	if err := ss.db.stormDB.Find("Status", status, &subscribes); err != nil {
		return nil, errors.Errorf("failed to find by status: %v", err)
	}

	return subscribes, nil
}

// Subscribe subscribes to newsletter
func (ss *subscribeService) Subscribe(token string) error {
	s, err := ss.FindByToken(token)
	if err != nil {
		return err
	}

	s.Status = blog.StatusSubscribed
	if err := ss.db.stormDB.Save(s); err != nil {
		return err
	}

	return nil
}

// Unsubscribe unsubscribes from newsletter
func (ss *subscribeService) Unsubscribe(email string) error {
	s, err := ss.FindByEmail(email)
	if err != nil {
		return errors.Errorf("failed to find by email: %v", err)
	}

	s.Status = blog.StatusUnsubscribed
	if err := ss.db.stormDB.Save(s); err != nil {
		return errors.Errorf("failed to save: %v", err)
	}

	return nil
}
