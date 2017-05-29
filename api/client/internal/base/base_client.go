// Package bclient is an internal package for the common API client functionality.
package bclient

/* distrilock - https://github.com/gdm85/distrilock
Copyright (C) 2017 gdm85
This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; either version 2 of the License, or
(at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

import (
	"bitbucket.org/gdm85/go-distrilock/api"
	"bitbucket.org/gdm85/go-distrilock/api/client"
)

type clientImpl interface {
	AcquireConn() error
	// Do is the function called to process a request on the wire and return the result.
	Do(req *api.LockRequest) (*api.LockResponse, error)
	// Close will release and close the underlying connection (if any).
	Close() error
}

type baseClient struct {
	clientImpl
}

func New(ci clientImpl) client.Client {
	return &baseClient{
		clientImpl: ci,
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

		return l, nil
	}

	return nil, &client.Error{Result: res.Result, Reason: res.Reason}
}

// Release will release a locked name previously acquired in this session.
func (c *baseClient) Release(l *client.Lock) error {
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

// Verify will verify that the lock is currently held by the client and healthy.
func (c *baseClient) Verify(l *client.Lock) error {
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
