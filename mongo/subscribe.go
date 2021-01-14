package mongo

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"

	"github.com/quantonganh/blog"
)

const (
	StatusPending      = "pending"
	StatusSubscribed   = "subscribed"
	StatusUnsubscribed = "unsubscribed"
)

type subscribeService struct {
	collection *mongo.Collection
}

func NewSubscribeService(db *DB) *subscribeService {
	newDB := db.NewDatabase("mailing_list")
	return &subscribeService{
		collection: newDB.Collection("subscribers"),
	}
}

func (ss *subscribeService) FindByEmail(email string) (*blog.Subscribe, error) {
	filter := bson.M{"email": email}

	s := blog.Subscribe{}
	err := ss.collection.FindOne(context.TODO(), filter).Decode(&s)

	return &s, err
}

func (ss *subscribeService) Insert(s *blog.Subscribe) error {
	_, err := ss.collection.InsertOne(context.TODO(), s)
	if err != nil {
		return errors.Errorf("failed to insert blog.Subscribe into the collection: %v", err)
	}

	return nil
}

func (ss *subscribeService) FindByToken(token string) (*blog.Subscribe, error) {
	filter := bson.M{"token": token}

	s := blog.Subscribe{}
	if err := ss.collection.FindOne(context.TODO(), filter).Decode(&s); err != nil {
		return nil, errors.Errorf("failed to decode blog.Subscribe into s: %v", err)
	}

	return &s, nil
}

func (ss *subscribeService) FindByStatus(status string) ([]blog.Subscribe, error) {
	filter := bson.M{"status": status}

	cursor, err := ss.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, errors.Errorf("failed to find blog.Subscribe by status %s: %v", status, err)
	}

	var subscribers []blog.Subscribe
	if err := cursor.All(context.TODO(), &subscribers); err != nil {
		return nil, errors.Errorf("failed to decode all matching subscribers: %v", err)
	}

	return subscribers, nil
}

func (ss *subscribeService) Subscribe(token string) error {
	filter := bson.M{"token": token}

	update := bson.M{
		"$set": bson.M{
			"status": StatusSubscribed,
		},
	}

	singleResult := ss.collection.FindOneAndUpdate(context.TODO(), filter, update)
	if err := singleResult.Err(); err != nil {
		return errors.Errorf("failed to update status of %s: %v", token, err)
	}

	return nil
}

func (ss *subscribeService) Unsubscribe(email string) error {
	filter := bson.M{"email": email}

	update := bson.M{
		"$set": bson.M{
			"status": StatusUnsubscribed,
		},
	}

	singleResult := ss.collection.FindOneAndUpdate(context.TODO(), filter, update)
	if err := singleResult.Err(); err != nil {
		return errors.Errorf("failed to update status of %s: %v", email, err)
	}

	return nil
}
