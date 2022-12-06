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
