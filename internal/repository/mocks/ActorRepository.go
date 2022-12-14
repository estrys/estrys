// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	context "context"

	models "github.com/estrys/estrys/internal/models"
	mock "github.com/stretchr/testify/mock"

	repository "github.com/estrys/estrys/internal/repository"

	url "net/url"
)

// ActorRepository is an autogenerated mock type for the ActorRepository type
type ActorRepository struct {
	mock.Mock
}

type ActorRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *ActorRepository) EXPECT() *ActorRepository_Expecter {
	return &ActorRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: _a0, _a1
func (_m *ActorRepository) Create(_a0 context.Context, _a1 repository.CreateActorRequest) (*models.Actor, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *models.Actor
	if rf, ok := ret.Get(0).(func(context.Context, repository.CreateActorRequest) *models.Actor); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Actor)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, repository.CreateActorRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ActorRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type ActorRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 repository.CreateActorRequest
func (_e *ActorRepository_Expecter) Create(_a0 interface{}, _a1 interface{}) *ActorRepository_Create_Call {
	return &ActorRepository_Create_Call{Call: _e.mock.On("Create", _a0, _a1)}
}

func (_c *ActorRepository_Create_Call) Run(run func(_a0 context.Context, _a1 repository.CreateActorRequest)) *ActorRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(repository.CreateActorRequest))
	})
	return _c
}

func (_c *ActorRepository_Create_Call) Return(_a0 *models.Actor, _a1 error) *ActorRepository_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Get provides a mock function with given fields: ctx, _a1
func (_m *ActorRepository) Get(ctx context.Context, _a1 *url.URL) (*models.Actor, error) {
	ret := _m.Called(ctx, _a1)

	var r0 *models.Actor
	if rf, ok := ret.Get(0).(func(context.Context, *url.URL) *models.Actor); ok {
		r0 = rf(ctx, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Actor)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *url.URL) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ActorRepository_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type ActorRepository_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 *url.URL
func (_e *ActorRepository_Expecter) Get(ctx interface{}, _a1 interface{}) *ActorRepository_Get_Call {
	return &ActorRepository_Get_Call{Call: _e.mock.On("Get", ctx, _a1)}
}

func (_c *ActorRepository_Get_Call) Run(run func(ctx context.Context, _a1 *url.URL)) *ActorRepository_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*url.URL))
	})
	return _c
}

func (_c *ActorRepository_Get_Call) Return(_a0 *models.Actor, _a1 error) *ActorRepository_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewActorRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewActorRepository creates a new instance of ActorRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewActorRepository(t mockConstructorTestingTNewActorRepository) *ActorRepository {
	mock := &ActorRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
