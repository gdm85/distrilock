package concurrent

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
	"sync"

	"github.com/gdm85/distrilock/api/client"
)

type concurrentWrapper struct {
	sync.Mutex
	c client.Client
}

// New returns a concurrency-safe client using the specified client.
func New(c client.Client) client.Client {
	if _, ok := c.(*concurrentWrapper); ok {
		panic("BUG: trying to wrap twice concurrency client")
	}
	return &concurrentWrapper{
		c: c,
	}
}

// Acquire will acquire a named lock through the distrilock daemon.
func (c *concurrentWrapper) Acquire(lockName string) (*client.Lock, error) {
	c.Lock()
	l, err := c.c.Acquire(lockName)
	c.Unlock()
	return l, err
}

// Release will release a locked name previously acquired in this session.
func (c *concurrentWrapper) Release(l *client.Lock) error {
	c.Lock()
	err := c.c.Release(l)
	c.Unlock()
	return err
}

// IsLocked returns true when distrilock deamon estabilished that lock is currently acquired.
func (c *concurrentWrapper) IsLocked(lockName string) (bool, error) {
	c.Lock()
	b, err := c.c.IsLocked(lockName)
	c.Unlock()
	return b, err
}

// Close will release all active locks and close the connection.
func (c *concurrentWrapper) Close() error {
	c.Lock()
	err := c.c.Close()
	c.Unlock()
	return err
}

// Verify will verify that the lock is currently held by the client and healthy.
func (c *concurrentWrapper) Verify(l *client.Lock) error {
	c.Lock()
	err := c.c.Verify(l)
	c.Unlock()
	return err
}
