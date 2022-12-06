// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	twitter "github.com/g8rswimmer/go-twitter/v2"
)

// Backend is an autogenerated mock type for the Backend type
type Backend struct {
	mock.Mock
}

// UserNameLookup provides a mock function with given fields: ctx, usernames, opts
func (_m *Backend) UserNameLookup(ctx context.Context, usernames []string, opts twitter.UserLookupOpts) (*twitter.UserLookupResponse, error) {
	ret := _m.Called(ctx, usernames, opts)

	var r0 *twitter.UserLookupResponse
	if rf, ok := ret.Get(0).(func(context.Context, []string, twitter.UserLookupOpts) *twitter.UserLookupResponse); ok {
		r0 = rf(ctx, usernames, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*twitter.UserLookupResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string, twitter.UserLookupOpts) error); ok {
		r1 = rf(ctx, usernames, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewBackend interface {
	mock.TestingT
	Cleanup(func())
}

// NewBackend creates a new instance of Backend. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewBackend(t mockConstructorTestingTNewBackend) *Backend {
	mock := &Backend{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
