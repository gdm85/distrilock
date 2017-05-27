package dlclientws

import (
	"encoding/gob"
	"fmt"
	"net"
	"time"

	"bitbucket.org/gdm85/go-distrilock/api"
	"bitbucket.org/gdm85/go-distrilock/api/client"
	
	"github.com/gorilla/websocket"
)

// Client is a single-connection, non-concurrency-safe client to a distrilock daemon.
type Client struct {
	endpoint                             string
	conn *websocket.Conn
	messageType int

	keepAlive, readTimeout, writeTimeout time.Duration
	locks map[*client.Lock]struct{}
}

// String returns a summary of the client connection and active locks.
func (c *Client) String() string {
	return fmt.Sprintf("%v with %d locks", c.conn, len(c.locks))
}

// NewBinary returns a new binary distrilock websocket client; no connection is performed.
func NewBinary(endpoint string, keepAlive, readTimeout, writeTimeout time.Duration) *Client {
	return &Client{
		endpoint:     endpoint,
		keepAlive:    keepAlive,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		locks:        map[*client.Lock]struct{}{},
		
		messageType : websocket.BinaryMessage,
	}
}

// NewJSON returns a new JSON distrilock websocket client; no connection is performed.
func NewJSON(endpoint string, keepAlive, readTimeout, writeTimeout time.Duration) *Client {
	return &Client{
		endpoint:     endpoint,
		keepAlive:    keepAlive,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		locks:        map[*client.Lock]struct{}{},
		
		messageType : websocket.TextMessage,
	}
}

// acquireConn is called every time a connection would be necessary; it does nothing if connection has already been made. It will re-estabilish a connection if Client c had been closed before.
func (c *Client) acquireConn() error {
	if c.conn == nil {
		var err error
		c.conn, _, err = websocket.DefaultDialer.Dial(c.endpoint, nil)
		if err != nil {
			return err
		}
		if c.keepAlive != 0 {
			uc := c.conn.UnderlyingConn()
			conn, ok := uc.(*net.TCPConn)
			if !ok {
				return fmt.Errorf("found connection type %T, but %T expected", uc, conn)
			}

			// setup keep-alive
			err = conn.SetKeepAlive(true)
			if err != nil {
				return err
			}
			err = conn.SetKeepAlivePeriod(c.keepAlive)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
