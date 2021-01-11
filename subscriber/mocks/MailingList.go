// Code generated by mockery v2.4.0. DO NOT EDIT.

package mocks

import (
	subscriber "github.com/quantonganh/blog/subscriber"
	mock "github.com/stretchr/testify/mock"
)

// MailingList is an autogenerated mock type for the MailingList type
type MailingList struct {
	mock.Mock
}

// FindByEmail provides a mock function with given fields: email
func (_m *MailingList) FindByEmail(email string) (*subscriber.Subscriber, error) {
	ret := _m.Called(email)

	var r0 *subscriber.Subscriber
	if rf, ok := ret.Get(0).(func(string) *subscriber.Subscriber); ok {
		r0 = rf(email)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*subscriber.Subscriber)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(email)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindByStatus provides a mock function with given fields: status
func (_m *MailingList) FindByStatus(status string) ([]subscriber.Subscriber, error) {
	ret := _m.Called(status)

	var r0 []subscriber.Subscriber
	if rf, ok := ret.Get(0).(func(string) []subscriber.Subscriber); ok {
		r0 = rf(status)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]subscriber.Subscriber)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(status)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindByToken provides a mock function with given fields: token
func (_m *MailingList) FindByToken(token string) (*subscriber.Subscriber, error) {
	ret := _m.Called(token)

	var r0 *subscriber.Subscriber
	if rf, ok := ret.Get(0).(func(string) *subscriber.Subscriber); ok {
		r0 = rf(token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*subscriber.Subscriber)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Insert provides a mock function with given fields: s
func (_m *MailingList) Insert(s *subscriber.Subscriber) error {
	ret := _m.Called(s)

	var r0 error
	if rf, ok := ret.Get(0).(func(*subscriber.Subscriber) error); ok {
		r0 = rf(s)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Subscribe provides a mock function with given fields: token
func (_m *MailingList) Subscribe(token string) error {
	ret := _m.Called(token)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(token)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Unsubscribe provides a mock function with given fields: email
func (_m *MailingList) Unsubscribe(email string) error {
	ret := _m.Called(email)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(email)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}