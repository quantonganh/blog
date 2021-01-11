package subscriber

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

const (
	StatusPending      = "pending"
	StatusSubscribed   = "subscribed"
	StatusUnsubscribed = "unsubscribed"
)

type mailingList struct {
	collection *mongo.Collection
}

func NewMailingList(db *mongo.Database) *mailingList {
	return &mailingList{
		collection: db.Collection("subscribers"),
	}
}

func (ml *mailingList) FindByEmail(email string) (*Subscriber, error) {
	filter := bson.M{"email": email}

	s := Subscriber{}
	err := ml.collection.FindOne(context.TODO(), filter).Decode(&s)

	return &s, err
}

func (ml *mailingList) Insert(s *Subscriber) error {
	_, err := ml.collection.InsertOne(context.TODO(), s)
	if err != nil {
		return errors.Errorf("failed to insert Subscriber into the collection: %v", err)
	}

	return nil
}

func (ml *mailingList) FindByToken(token string) (*Subscriber, error) {
	filter := bson.M{"token": token}

	s := Subscriber{}
	if err := ml.collection.FindOne(context.TODO(), filter).Decode(&s); err != nil {
		return nil, errors.Errorf("failed to decode Subscriber into s: %v", err)
	}

	return &s, nil
}

func (ml *mailingList) FindByStatus(status string) ([]Subscriber, error) {
	filter := bson.M{"status": status}

	cursor, err := ml.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, errors.Errorf("failed to find Subscriber by status %s: %v", status, err)
	}

	var subscribers []Subscriber
	if err := cursor.All(context.TODO(), &subscribers); err != nil {
		return nil, errors.Errorf("failed to decode all matching subscribers: %v", err)
	}

	return subscribers, nil
}

func (ml *mailingList) Subscribe(token string) error {
	filter := bson.M{"token": token}

	update := bson.M{
		"$set": bson.M{
			"status": StatusSubscribed,
		},
	}

	singleResult := ml.collection.FindOneAndUpdate(context.TODO(), filter, update)
	if err := singleResult.Err(); err != nil {
		return errors.Errorf("failed to update status of %s: %v", token, err)
	}

	return nil
}

func (ml *mailingList) Unsubscribe(email string) error {
	filter := bson.M{"email": email}

	update := bson.M{
		"$set": bson.M{
			"status": StatusUnsubscribed,
		},
	}

	singleResult := ml.collection.FindOneAndUpdate(context.TODO(), filter, update)
	if err := singleResult.Err(); err != nil {
		return errors.Errorf("failed to update status of %s: %v", email, err)
	}

	return nil
}
