// Code generated by mockery v2.4.0. DO NOT EDIT.

package mock

import (
	blog "github.com/quantonganh/blog"
	mock "github.com/stretchr/testify/mock"
)

// SMTPService is an autogenerated mock type for the SMTPService type
type SMTPService struct {
	mock.Mock
}

// GenerateNewUUID provides a mock function with given fields:
func (_m *SMTPService) GenerateNewUUID() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetHMACSecret provides a mock function with given fields:
func (_m *SMTPService) GetHMACSecret() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// SendConfirmationEmail provides a mock function with given fields: to, token
func (_m *SMTPService) SendConfirmationEmail(to string, token string) error {
	ret := _m.Called(to, token)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(to, token)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendNewsletter provides a mock function with given fields: posts
func (_m *SMTPService) SendNewsletter(posts []*blog.Post) {
	_m.Called(posts)
}

// SendThankYouEmail provides a mock function with given fields: to
func (_m *SMTPService) SendThankYouEmail(to string) error {
	ret := _m.Called(to)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(to)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
