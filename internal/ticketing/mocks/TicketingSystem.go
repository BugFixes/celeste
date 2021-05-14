// Code generated by mockery 2.7.5. DO NOT EDIT.

package mocks

import (
	ticketing "github.com/bugfixes/celeste/internal/ticketing"
	mock "github.com/stretchr/testify/mock"
)

// TicketingSystem is an autogenerated mock type for the TicketingSystem type
type TicketingSystem struct {
	mock.Mock
}

// Connect provides a mock function with given fields:
func (_m *TicketingSystem) Connect() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Create provides a mock function with given fields: _a0
func (_m *TicketingSystem) Create(_a0 ticketing.Ticket) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(ticketing.Ticket) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Fetch provides a mock function with given fields: _a0
func (_m *TicketingSystem) Fetch(_a0 ticketing.Ticket) (ticketing.Ticket, error) {
	ret := _m.Called(_a0)

	var r0 ticketing.Ticket
	if rf, ok := ret.Get(0).(func(ticketing.Ticket) ticketing.Ticket); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(ticketing.Ticket)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ticketing.Ticket) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FetchRemoteTicket provides a mock function with given fields: _a0
func (_m *TicketingSystem) FetchRemoteTicket(_a0 ticketing.Hash) (ticketing.Ticket, error) {
	ret := _m.Called(_a0)

	var r0 ticketing.Ticket
	if rf, ok := ret.Get(0).(func(ticketing.Hash) ticketing.Ticket); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(ticketing.Ticket)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ticketing.Hash) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ParseCredentials provides a mock function with given fields: _a0
func (_m *TicketingSystem) ParseCredentials(_a0 interface{}) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: _a0
func (_m *TicketingSystem) Update(_a0 ticketing.Ticket) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(ticketing.Ticket) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
