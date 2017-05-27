package dlclient

import (
	"encoding/gob"
	"fmt"
	"net"
	"time"

	"bitbucket.org/gdm85/go-distrilock/api"
	"bitbucket.org/gdm85/go-distrilock/api/client"
	"bitbucket.org/gdm85/go-distrilock/api/client/internal/base"
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

// New returns a new distrilock client; no connection is performed.
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
