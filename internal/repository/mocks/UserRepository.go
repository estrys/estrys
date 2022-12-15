// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	context "context"

	models "github.com/estrys/estrys/internal/models"
	mock "github.com/stretchr/testify/mock"

	repository "github.com/estrys/estrys/internal/repository"
)

// UserRepository is an autogenerated mock type for the UserRepository type
type UserRepository struct {
	mock.Mock
}

type UserRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *UserRepository) EXPECT() *UserRepository_Expecter {
	return &UserRepository_Expecter{mock: &_m.Mock}
}

// CreateUser provides a mock function with given fields: _a0, _a1
func (_m *UserRepository) CreateUser(_a0 context.Context, _a1 repository.CreateUserRequest) (*models.User, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *models.User
	if rf, ok := ret.Get(0).(func(context.Context, repository.CreateUserRequest) *models.User); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, repository.CreateUserRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserRepository_CreateUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateUser'
type UserRepository_CreateUser_Call struct {
	*mock.Call
}

// CreateUser is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 repository.CreateUserRequest
func (_e *UserRepository_Expecter) CreateUser(_a0 interface{}, _a1 interface{}) *UserRepository_CreateUser_Call {
	return &UserRepository_CreateUser_Call{Call: _e.mock.On("CreateUser", _a0, _a1)}
}

func (_c *UserRepository_CreateUser_Call) Run(run func(_a0 context.Context, _a1 repository.CreateUserRequest)) *UserRepository_CreateUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(repository.CreateUserRequest))
	})
	return _c
}

func (_c *UserRepository_CreateUser_Call) Return(_a0 *models.User, _a1 error) *UserRepository_CreateUser_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Follow provides a mock function with given fields: _a0, _a1, _a2
func (_m *UserRepository) Follow(_a0 context.Context, _a1 *models.User, _a2 *models.Actor) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.User, *models.Actor) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UserRepository_Follow_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Follow'
type UserRepository_Follow_Call struct {
	*mock.Call
}

// Follow is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *models.User
//   - _a2 *models.Actor
func (_e *UserRepository_Expecter) Follow(_a0 interface{}, _a1 interface{}, _a2 interface{}) *UserRepository_Follow_Call {
	return &UserRepository_Follow_Call{Call: _e.mock.On("Follow", _a0, _a1, _a2)}
}

func (_c *UserRepository_Follow_Call) Run(run func(_a0 context.Context, _a1 *models.User, _a2 *models.Actor)) *UserRepository_Follow_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*models.User), args[2].(*models.Actor))
	})
	return _c
}

func (_c *UserRepository_Follow_Call) Return(_a0 error) *UserRepository_Follow_Call {
	_c.Call.Return(_a0)
	return _c
}

// Get provides a mock function with given fields: _a0, _a1
func (_m *UserRepository) Get(_a0 context.Context, _a1 string) (*models.User, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *models.User
	if rf, ok := ret.Get(0).(func(context.Context, string) *models.User); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserRepository_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type UserRepository_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
func (_e *UserRepository_Expecter) Get(_a0 interface{}, _a1 interface{}) *UserRepository_Get_Call {
	return &UserRepository_Get_Call{Call: _e.mock.On("Get", _a0, _a1)}
}

func (_c *UserRepository_Get_Call) Run(run func(_a0 context.Context, _a1 string)) *UserRepository_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *UserRepository_Get_Call) Return(_a0 *models.User, _a1 error) *UserRepository_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetFollowers provides a mock function with given fields: _a0, _a1
func (_m *UserRepository) GetFollowers(_a0 context.Context, _a1 *models.User) (models.ActorSlice, error) {
	ret := _m.Called(_a0, _a1)

	var r0 models.ActorSlice
	if rf, ok := ret.Get(0).(func(context.Context, *models.User) models.ActorSlice); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(models.ActorSlice)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *models.User) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserRepository_GetFollowers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFollowers'
type UserRepository_GetFollowers_Call struct {
	*mock.Call
}

// GetFollowers is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *models.User
func (_e *UserRepository_Expecter) GetFollowers(_a0 interface{}, _a1 interface{}) *UserRepository_GetFollowers_Call {
	return &UserRepository_GetFollowers_Call{Call: _e.mock.On("GetFollowers", _a0, _a1)}
}

func (_c *UserRepository_GetFollowers_Call) Run(run func(_a0 context.Context, _a1 *models.User)) *UserRepository_GetFollowers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*models.User))
	})
	return _c
}

func (_c *UserRepository_GetFollowers_Call) Return(_a0 models.ActorSlice, _a1 error) *UserRepository_GetFollowers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetWithFollowers provides a mock function with given fields: ctx
func (_m *UserRepository) GetWithFollowers(ctx context.Context) (models.UserSlice, error) {
	ret := _m.Called(ctx)

	var r0 models.UserSlice
	if rf, ok := ret.Get(0).(func(context.Context) models.UserSlice); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(models.UserSlice)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserRepository_GetWithFollowers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetWithFollowers'
type UserRepository_GetWithFollowers_Call struct {
	*mock.Call
}

// GetWithFollowers is a helper method to define mock.On call
//   - ctx context.Context
func (_e *UserRepository_Expecter) GetWithFollowers(ctx interface{}) *UserRepository_GetWithFollowers_Call {
	return &UserRepository_GetWithFollowers_Call{Call: _e.mock.On("GetWithFollowers", ctx)}
}

func (_c *UserRepository_GetWithFollowers_Call) Run(run func(ctx context.Context)) *UserRepository_GetWithFollowers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *UserRepository_GetWithFollowers_Call) Return(_a0 models.UserSlice, _a1 error) *UserRepository_GetWithFollowers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// UnFollow provides a mock function with given fields: _a0, _a1, _a2
func (_m *UserRepository) UnFollow(_a0 context.Context, _a1 *models.User, _a2 *models.Actor) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.User, *models.Actor) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UserRepository_UnFollow_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UnFollow'
type UserRepository_UnFollow_Call struct {
	*mock.Call
}

// UnFollow is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *models.User
//   - _a2 *models.Actor
func (_e *UserRepository_Expecter) UnFollow(_a0 interface{}, _a1 interface{}, _a2 interface{}) *UserRepository_UnFollow_Call {
	return &UserRepository_UnFollow_Call{Call: _e.mock.On("UnFollow", _a0, _a1, _a2)}
}

func (_c *UserRepository_UnFollow_Call) Run(run func(_a0 context.Context, _a1 *models.User, _a2 *models.Actor)) *UserRepository_UnFollow_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*models.User), args[2].(*models.Actor))
	})
	return _c
}

func (_c *UserRepository_UnFollow_Call) Return(_a0 error) *UserRepository_UnFollow_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewUserRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewUserRepository creates a new instance of UserRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUserRepository(t mockConstructorTestingTNewUserRepository) *UserRepository {
	mock := &UserRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
