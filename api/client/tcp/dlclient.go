package dlclient

import (
	"encoding/gob"
	"fmt"
	"net"
	"time"

	"bitbucket.org/gdm85/go-distrilock/api"
)

// Client is a single-connection, non-concurrency-safe client to a distrilock daemon.
type Client struct {
	endpoint                             *net.TCPAddr
	keepAlive, readTimeout, writeTimeout time.Duration

	conn *net.TCPConn
	d    *gob.Decoder
	e    *gob.Encoder

	locks map[*Lock]struct{}
}

// String returns a summary of the client connection and active locks.
func (c *Client) String() string {
	return fmt.Sprintf("%v with %d locks", c.conn, len(c.locks))
}

// ClientError is the composite error return by all client method calls.
type ClientError struct {
	Result api.LockCommandResult
	Reason string
}

// Error returns the associated summary of the ClientError e.
func (e *ClientError) Error() string {
	return fmt.Sprintf("%v: %s", e.Result, e.Reason)
}

// Lock is a client-specific acquired lock object.
type Lock struct {
	c    *Client
	name string
}

// String returns the lock name and the associated client.
func (l *Lock) String() string {
	return fmt.Sprintf("%s on %v", l.name, l.c)
}

// New returns a new distrilock client; no connection is performed.
func New(endpoint *net.TCPAddr, keepAlive, readTimeout, writeTimeout time.Duration) *Client {
	return &Client{
		endpoint:     endpoint,
		keepAlive:    keepAlive,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		locks:        map[*Lock]struct{}{},
	}
}

// acquireConn is called every time a connection would be necessary; it does nothing if connection has already been made. It will re-estabilish a connection if Client c had been closed before.
func (c *Client) acquireConn() error {
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

// Acquire will acquire a named lock through the distrilock daemon.
func (c *Client) Acquire(lockName string) (*Lock, error) {
	err := c.acquireConn()
	if err != nil {
		return nil, err
	}

	var req api.LockRequest
	req.VersionMajor, req.VersionMinor = api.VersionMajor, api.VersionMinor
	req.Command = api.Acquire
	req.LockName = lockName

	res, err := c.do(&req)
	if err != nil {
		return nil, err
	}

	if res.Result == api.Success {
		// create lock and return it
		l := &Lock{
			c:    c,
			name: lockName,
		}

		c.locks[l] = struct{}{}

		return l, nil
	}

	return nil, &ClientError{Result: res.Result, Reason: res.Reason}
}

// do is the function called to process a request on the wire and return the result.
func (c *Client) do(req *api.LockRequest) (*api.LockResponse, error) {
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

// Release will release a locked name previously acquired.
func (l *Lock) Release() error {
	err := l.c.acquireConn()
	if err != nil {
		return err
	}

	var req api.LockRequest
	req.VersionMajor, req.VersionMinor = api.VersionMajor, api.VersionMinor
	req.Command = api.Release
	req.LockName = l.name

	res, err := l.c.do(&req)
	if err != nil {
		return err
	}

	if res.Result == api.Success {
		delete(l.c.locks, l)
		return nil
	}

	return &ClientError{Result: res.Result, Reason: res.Reason}
}

// IsLocked returns true when distrilock deamon estabilished that lock is currently acquired.
func (c *Client) IsLocked(lockName string) (bool, error) {
	err := c.acquireConn()
	if err != nil {
		return false, err
	}

	var req api.LockRequest
	req.VersionMajor, req.VersionMinor = api.VersionMajor, api.VersionMinor
	req.Command = api.Peek
	req.LockName = lockName

	res, err := c.do(&req)
	if err != nil {
		return false, err
	}

	if res.Result == api.Success {
		return res.IsLocked, nil
	}

	return false, &ClientError{Result: res.Result, Reason: res.Reason}
}

// Close will release all active locks and close the connection.
func (c *Client) Close() error {
	if c.conn != nil {
		// explicitly release all locks
		for l := range c.locks {
			// ignore release errors
			l.Release()
		}

		err := c.conn.Close()
		c.conn = nil
		c.d = nil
		c.e = nil
		return err
	}
	return nil
}

// Verify will verify that the lock is currently held by the client and healthy.
func (l *Lock) Verify() error {
	err := l.c.acquireConn()
	if err != nil {
		return err
	}

	var req api.LockRequest
	req.VersionMajor, req.VersionMinor = api.VersionMajor, api.VersionMinor
	req.Command = api.Verify
	req.LockName = l.name

	res, err := l.c.do(&req)
	if err != nil {
		return err
	}

	if res.Result == api.Success {
		return nil
	}

	return &ClientError{Result: res.Result, Reason: res.Reason}
}
