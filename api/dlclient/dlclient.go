package dlclient

import (
	"encoding/gob"
	"fmt"
	"net"
	"time"

	"bitbucket.org/gdm85/go-distrilock/api"
)

type Client struct {
	endpoint                             *net.TCPAddr
	keepAlive, readTimeout, writeTimeout time.Duration

	conn *net.TCPConn
	d    *gob.Decoder
	e    *gob.Encoder

	locks map[*Lock]struct{}
}

type Lock struct {
	c    *Client
	name string
}

func New(endpoint *net.TCPAddr, keepAlive, readTimeout, writeTimeout time.Duration) *Client {
	return &Client{
		endpoint:     endpoint,
		keepAlive:    keepAlive,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		locks:        map[*Lock]struct{}{},
	}
}

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

	return nil, fmt.Errorf("%v: %s", res.Result, res.Reason)
}

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

	return fmt.Errorf("%v: %s", res.Result, res.Reason)
}

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

	return false, fmt.Errorf("%v: %s", res.Result, res.Reason)
}

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
