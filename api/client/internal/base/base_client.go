package bclient

import (
	"fmt"

	"bitbucket.org/gdm85/go-distrilock/api"
	"bitbucket.org/gdm85/go-distrilock/api/client"
)

type baseClient struct {
	client.ClientImpl
	
	locks map[*client.Lock]struct{}
}

// String returns a summary of the client connection and active locks.
func (c *baseClient) String() string {
	return fmt.Sprintf("%v with %d locks", c.ClientImpl, len(c.locks))
}

func New(ci client.ClientImpl) client.Client {
	return &baseClient{
		ClientImpl: ci,
	}
}

// Acquire will acquire a named lock through the distrilock daemon.
func (c *baseClient) Acquire(lockName string) (*client.Lock, error) {
	err := c.AcquireConn()
	if err != nil {
		return nil, err
	}

	var req api.LockRequest
	req.VersionMajor, req.VersionMinor = api.VersionMajor, api.VersionMinor
	req.Command = api.Acquire
	req.LockName = lockName

	res, err := c.Do(&req)
	if err != nil {
		return nil, err
	}

	if res.Result == api.Success {
		// create lock and return it
		l := &client.Lock{Client: c, Name: lockName}

		c.locks[l] = struct{}{}

		return l, nil
	}

	return nil, &client.Error{Result: res.Result, Reason: res.Reason}
}

// Release will release a locked name previously acquired.
func (c *baseClient) Release(l *client.Lock) error {
	if c != l.Client {
		panic("BUG: attempting to release lock acquired via different client")
	}
	err := c.AcquireConn()
	if err != nil {
		return err
	}

	var req api.LockRequest
	req.VersionMajor, req.VersionMinor = api.VersionMajor, api.VersionMinor
	req.Command = api.Release
	req.LockName = l.Name

	res, err := c.Do(&req)
	if err != nil {
		return err
	}

	if res.Result == api.Success {
		delete(c.locks, l)
		return nil
	}

	return &client.Error{Result: res.Result, Reason: res.Reason}
}

// IsLocked returns true when distrilock deamon estabilished that lock is currently acquired.
func (c *baseClient) IsLocked(lockName string) (bool, error) {
	err := c.AcquireConn()
	if err != nil {
		return false, err
	}

	var req api.LockRequest
	req.VersionMajor, req.VersionMinor = api.VersionMajor, api.VersionMinor
	req.Command = api.Peek
	req.LockName = lockName

	res, err := c.Do(&req)
	if err != nil {
		return false, err
	}

	if res.Result == api.Success {
		return res.IsLocked, nil
	}

	return false, &client.Error{Result: res.Result, Reason: res.Reason}
}

// Close will release all active locks and close the connection.
func (c *baseClient) Close() error {
	// explicitly release all locks
	for l := range c.locks {
		// ignore release errors
		l.Release()
	}

	return c.ClientImpl.Close()
}

// Verify will verify that the lock is currently held by the client and healthy.
func (c *baseClient) Verify(l *client.Lock) error {
	if c != l.Client {
		panic("BUG: attempting to release lock acquired via different client")
	}
	err := c.AcquireConn()
	if err != nil {
		return err
	}

	var req api.LockRequest
	req.VersionMajor, req.VersionMinor = api.VersionMajor, api.VersionMinor
	req.Command = api.Verify
	req.LockName = l.Name

	res, err := c.Do(&req)
	if err != nil {
		return err
	}

	if res.Result == api.Success {
		return nil
	}

	return &client.Error{Result: res.Result, Reason: res.Reason}
}
