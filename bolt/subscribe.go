package bolt

import (
	"github.com/go-errors/errors"

	"github.com/quantonganh/blog"
)

func (db *DB) FindByEmail(email string) (*blog.Subscribe, error) {
	var s blog.Subscribe
	if err := db.db.One("Email", email, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (db *DB) Insert(s *blog.Subscribe) error {
	if err := db.db.Save(s); err != nil {
		return errors.Errorf("failed to save: %v", err)
	}

	return nil
}

func (db *DB) FindByToken(token string) (*blog.Subscribe, error) {
	var s blog.Subscribe
	if err := db.db.One("Token", token, &s); err != nil {
		return nil, errors.Errorf("failed to find by token: %v", err)
	}

	return &s, nil
}

func (db *DB) FindByStatus(status string) ([]blog.Subscribe, error) {
	var subscribes []blog.Subscribe
	if err := db.db.Find("Status", status, &subscribes); err != nil {
		return nil, errors.Errorf("failed to find by status: %v", err)
	}

	return subscribes, nil
}

func (db *DB) Subscribe(token string) error {
	s, err := db.FindByToken(token)
	if err != nil {
		return err
	}

	s.Status = blog.StatusSubscribed
	if err := db.db.Save(s); err != nil {
		return err
	}

	return nil
}

func (db *DB) Unsubscribe(email string) error {
	s, err := db.FindByEmail(email)
	if err != nil {
		return errors.Errorf("failed to find by email: %v", err)
	}

	s.Status = blog.StatusUnsubscribed
	if err := db.db.Save(s); err != nil {
		return errors.Errorf("failed to save: %v", err)
	}

	return nil
}
