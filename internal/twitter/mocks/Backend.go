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

type Backend_Expecter struct {
	mock *mock.Mock
}

func (_m *Backend) EXPECT() *Backend_Expecter {
	return &Backend_Expecter{mock: &_m.Mock}
}

// TweetLookup provides a mock function with given fields: ctx, ids, opts
func (_m *Backend) TweetLookup(ctx context.Context, ids []string, opts twitter.TweetLookupOpts) (*twitter.TweetLookupResponse, error) {
	ret := _m.Called(ctx, ids, opts)

	var r0 *twitter.TweetLookupResponse
	if rf, ok := ret.Get(0).(func(context.Context, []string, twitter.TweetLookupOpts) *twitter.TweetLookupResponse); ok {
		r0 = rf(ctx, ids, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*twitter.TweetLookupResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string, twitter.TweetLookupOpts) error); ok {
		r1 = rf(ctx, ids, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Backend_TweetLookup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TweetLookup'
type Backend_TweetLookup_Call struct {
	*mock.Call
}

// TweetLookup is a helper method to define mock.On call
//   - ctx context.Context
//   - ids []string
//   - opts twitter.TweetLookupOpts
func (_e *Backend_Expecter) TweetLookup(ctx interface{}, ids interface{}, opts interface{}) *Backend_TweetLookup_Call {
	return &Backend_TweetLookup_Call{Call: _e.mock.On("TweetLookup", ctx, ids, opts)}
}

func (_c *Backend_TweetLookup_Call) Run(run func(ctx context.Context, ids []string, opts twitter.TweetLookupOpts)) *Backend_TweetLookup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]string), args[2].(twitter.TweetLookupOpts))
	})
	return _c
}

func (_c *Backend_TweetLookup_Call) Return(_a0 *twitter.TweetLookupResponse, _a1 error) *Backend_TweetLookup_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// UserLookup provides a mock function with given fields: ctx, ids, opts
func (_m *Backend) UserLookup(ctx context.Context, ids []string, opts twitter.UserLookupOpts) (*twitter.UserLookupResponse, error) {
	ret := _m.Called(ctx, ids, opts)

	var r0 *twitter.UserLookupResponse
	if rf, ok := ret.Get(0).(func(context.Context, []string, twitter.UserLookupOpts) *twitter.UserLookupResponse); ok {
		r0 = rf(ctx, ids, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*twitter.UserLookupResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string, twitter.UserLookupOpts) error); ok {
		r1 = rf(ctx, ids, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Backend_UserLookup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UserLookup'
type Backend_UserLookup_Call struct {
	*mock.Call
}

// UserLookup is a helper method to define mock.On call
//   - ctx context.Context
//   - ids []string
//   - opts twitter.UserLookupOpts
func (_e *Backend_Expecter) UserLookup(ctx interface{}, ids interface{}, opts interface{}) *Backend_UserLookup_Call {
	return &Backend_UserLookup_Call{Call: _e.mock.On("UserLookup", ctx, ids, opts)}
}

func (_c *Backend_UserLookup_Call) Run(run func(ctx context.Context, ids []string, opts twitter.UserLookupOpts)) *Backend_UserLookup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]string), args[2].(twitter.UserLookupOpts))
	})
	return _c
}

func (_c *Backend_UserLookup_Call) Return(_a0 *twitter.UserLookupResponse, _a1 error) *Backend_UserLookup_Call {
	_c.Call.Return(_a0, _a1)
	return _c
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

// Backend_UserNameLookup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UserNameLookup'
type Backend_UserNameLookup_Call struct {
	*mock.Call
}

// UserNameLookup is a helper method to define mock.On call
//   - ctx context.Context
//   - usernames []string
//   - opts twitter.UserLookupOpts
func (_e *Backend_Expecter) UserNameLookup(ctx interface{}, usernames interface{}, opts interface{}) *Backend_UserNameLookup_Call {
	return &Backend_UserNameLookup_Call{Call: _e.mock.On("UserNameLookup", ctx, usernames, opts)}
}

func (_c *Backend_UserNameLookup_Call) Run(run func(ctx context.Context, usernames []string, opts twitter.UserLookupOpts)) *Backend_UserNameLookup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]string), args[2].(twitter.UserLookupOpts))
	})
	return _c
}

func (_c *Backend_UserNameLookup_Call) Return(_a0 *twitter.UserLookupResponse, _a1 error) *Backend_UserNameLookup_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// UserTweetTimeline provides a mock function with given fields: ctx, userID, opts
func (_m *Backend) UserTweetTimeline(ctx context.Context, userID string, opts twitter.UserTweetTimelineOpts) (*twitter.UserTweetTimelineResponse, error) {
	ret := _m.Called(ctx, userID, opts)

	var r0 *twitter.UserTweetTimelineResponse
	if rf, ok := ret.Get(0).(func(context.Context, string, twitter.UserTweetTimelineOpts) *twitter.UserTweetTimelineResponse); ok {
		r0 = rf(ctx, userID, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*twitter.UserTweetTimelineResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, twitter.UserTweetTimelineOpts) error); ok {
		r1 = rf(ctx, userID, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Backend_UserTweetTimeline_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UserTweetTimeline'
type Backend_UserTweetTimeline_Call struct {
	*mock.Call
}

// UserTweetTimeline is a helper method to define mock.On call
//   - ctx context.Context
//   - userID string
//   - opts twitter.UserTweetTimelineOpts
func (_e *Backend_Expecter) UserTweetTimeline(ctx interface{}, userID interface{}, opts interface{}) *Backend_UserTweetTimeline_Call {
	return &Backend_UserTweetTimeline_Call{Call: _e.mock.On("UserTweetTimeline", ctx, userID, opts)}
}

func (_c *Backend_UserTweetTimeline_Call) Run(run func(ctx context.Context, userID string, opts twitter.UserTweetTimelineOpts)) *Backend_UserTweetTimeline_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(twitter.UserTweetTimelineOpts))
	})
	return _c
}

func (_c *Backend_UserTweetTimeline_Call) Return(_a0 *twitter.UserTweetTimelineResponse, _a1 error) *Backend_UserTweetTimeline_Call {
	_c.Call.Return(_a0, _a1)
	return _c
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
