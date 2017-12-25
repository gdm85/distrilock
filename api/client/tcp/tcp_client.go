// Package tcp provides a distrilock client over TCP.
package tcp

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
	"encoding/gob"
	"fmt"
	"net"
	"time"

	"github.com/gdm85/distrilock/api"
	"github.com/gdm85/distrilock/api/client"
	"github.com/gdm85/distrilock/api/client/internal/base"
)

// Client is a single-connection, non-concurrency-safe client to a distrilock daemon.
type tcpClient struct {
	endpoint *net.TCPAddr
	conn     *net.TCPConn
	d        *gob.Decoder
	e        *gob.Encoder

	keepAlive, readTimeout, writeTimeout time.Duration
}

// String returns a summary of the client connection and active locks.
func (c *tcpClient) String() string {
	return fmt.Sprintf("%v", c.conn)
}

// New returns a new distrilock client; no connection is performed until the client is actually used.
func New(endpoint *net.TCPAddr, keepAlive, readTimeout, writeTimeout time.Duration) client.Client {
	return bclient.New(&tcpClient{
		endpoint:     endpoint,
		keepAlive:    keepAlive,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	})
}

// acquireConn is called every time a connection would be necessary; it does nothing if connection has already been made. It will re-estabilish a connection if Client c had been closed before.
func (c *tcpClient) AcquireConn() error {
	if c.conn == nil {
		var err error
		c.conn, err = net.DialTCP("tcp", nil, c.endpoint)
		if err != nil {
			return err
		}
		if c.keepAlive != 0 {
			// setup keep-alive
			err = c.conn.SetKeepAlive(true)
			if err != nil {
				return err
			}
			err = c.conn.SetKeepAlivePeriod(c.keepAlive)
			if err != nil {
				return err
			}
		}
		c.d = gob.NewDecoder(c.conn)
		c.e = gob.NewEncoder(c.conn)
	}
	return nil
}

func (c *tcpClient) Do(req *api.LockRequest) (*api.LockResponse, error) {
	if c.writeTimeout != 0 {
		err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
		if err != nil {
			return nil, err
		}
	}

	err := c.e.Encode(&req)
	if err != nil {
		return nil, err
	}

	// wait for a response
	var res api.LockResponse
	if c.readTimeout != 0 {
		err := c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
		if err != nil {
			return nil, err
		}
	}
	err = c.d.Decode(&res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *tcpClient) Close() error {
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	if err != nil {
		return err
	}
	c.conn = nil
	c.d = nil
	c.e = nil

	return nil
}
