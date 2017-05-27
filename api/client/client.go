package client

import (
	"fmt"

	"bitbucket.org/gdm85/go-distrilock/api"
)

type Client interface {
	Acquire(lockName string) (*Lock, error)
	Release(l *Lock) error
	IsLocked(lockName string) (bool, error)
	Verify(l *Lock) error
	Close() error
}

type ClientImpl interface {
	AcquireConn() error
	Do(req *api.LockRequest) (*api.LockResponse, error)
	Close() error
}

// Error is the composite error return by all client method calls.
type Error struct {
	Result api.LockCommandResult
	Reason string
}

// Error returns the associated summary of the ClientError e.
func (e *Error) Error() string {
	return fmt.Sprintf("%v: %s", e.Result, e.Reason)
}

// Lock is a client-specific acquired lock object.
type Lock struct {
	Client
	Name string
}

// String returns the lock name and the associated client.
func (l *Lock) String() string {
	return fmt.Sprintf("%s on %v", l.Name, l.Client)
}

// Release is a short-hand to call Client.Release for Lock l.
func (l *Lock) Release() error {
	return l.Client.Release(l)
}

// Verify is a short-hand to call Client.Verify for Lock l.
func (l *Lock) Verify() error {
	return l.Client.Verify(l)
}
