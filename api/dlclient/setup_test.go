package dlclient

import (
	"net"
	"os"
	"testing"
	"time"
)

const (
	defaultServerA = ":63419"
	defaultServerB = ":63420"
	defaultServerC = ":63421"
	defaultServerD = "sibling:63422"
)

var (
	testClientA1, testClientA2 *Client
	testClientB1               *Client
	testClientC1               *Client
	testClientD1               *Client
)

func init() {
	// first server process
	a, err := net.ResolveTCPAddr("tcp", defaultServerA)
	if err != nil {
		panic(err)
	}
	// a second process accessing same locks
	b, err := net.ResolveTCPAddr("tcp", defaultServerB)
	if err != nil {
		panic(err)
	}

	// first process on NFS
	c, err := net.ResolveTCPAddr("tcp", defaultServerC)
	if err != nil {
		panic(err)
	}
	// second process on NFS, different machine
	d, err := net.ResolveTCPAddr("tcp", defaultServerD)
	if err != nil {
		panic(err)
	}

	testClientA1 = New(a, time.Second*3, time.Second*2, time.Second*2)
	testClientA2 = New(a, time.Second*3, time.Second*2, time.Second*2)
	testClientB1 = New(b, time.Second*3, time.Second*2, time.Second*2)
	testClientC1 = New(c, time.Second*3, time.Second*2, time.Second*2)
	testClientD1 = New(d, time.Second*3, time.Second*2, time.Second*2)
}

func TestMain(m *testing.M) {
	retCode := m.Run()

	os.Exit(retCode)
}
