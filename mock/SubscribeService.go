// Code generated by mockery v2.4.0. DO NOT EDIT.

package mock

import (
	blog "github.com/quantonganh/blog"
	mock "github.com/stretchr/testify/mock"
)

// SubscribeService is an autogenerated mock type for the SubscribeService type
type SubscribeService struct {
	mock.Mock
}

// FindByEmail provides a mock function with given fields: email
func (_m *SubscribeService) FindByEmail(email string) (*blog.Subscribe, error) {
	ret := _m.Called(email)

	var r0 *blog.Subscribe
	if rf, ok := ret.Get(0).(func(string) *blog.Subscribe); ok {
		r0 = rf(email)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blog.Subscribe)
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
func (_m *SubscribeService) FindByStatus(status string) ([]blog.Subscribe, error) {
	ret := _m.Called(status)

	var r0 []blog.Subscribe
	if rf, ok := ret.Get(0).(func(string) []blog.Subscribe); ok {
		r0 = rf(status)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]blog.Subscribe)
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
func (_m *SubscribeService) FindByToken(token string) (*blog.Subscribe, error) {
	ret := _m.Called(token)

	var r0 *blog.Subscribe
	if rf, ok := ret.Get(0).(func(string) *blog.Subscribe); ok {
		r0 = rf(token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*blog.Subscribe)
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
func (_m *SubscribeService) Insert(s *blog.Subscribe) error {
	ret := _m.Called(s)

	var r0 error
	if rf, ok := ret.Get(0).(func(*blog.Subscribe) error); ok {
		r0 = rf(s)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Subscribe provides a mock function with given fields: token
func (_m *SubscribeService) Subscribe(token string) error {
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
func (_m *SubscribeService) Unsubscribe(email string) error {
	ret := _m.Called(email)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(email)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateStatus provides a mock function with given fields: email
func (_m *SubscribeService) UpdateStatus(email string) error {
	ret := _m.Called(email)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(email)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
