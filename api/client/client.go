package client

import (
	"fmt"

	"bitbucket.org/gdm85/go-distrilock/api"
)

// Client is a generic distrilock client interface.
type Client interface {
	// Acquire will acquire a named lock through the distrilock daemon.
	Acquire(lockName string) (*Lock, error)
	// Release will release a locked name previously acquired in this session.
	Release(l *Lock) error
	// IsLocked returns true when distrilock deamon estabilished that lock is currently acquired.
	IsLocked(lockName string) (bool, error)
	// Verify will verify that the lock is currently held by the client and healthy.
	Verify(l *Lock) error
	// Close releases all session-specific resources of this client.
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
