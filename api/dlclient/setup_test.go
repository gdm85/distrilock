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
)

var (
	testClientA1, testClientA2 *Client
)

func init() {
	a, err := net.ResolveTCPAddr("tcp", defaultServerA)
	if err != nil {
		panic(err)
	}

	testClientA1 = New(a, time.Second*3, time.Second*2, time.Second*2)
	testClientA2 = New(a, time.Second*3, time.Second*2, time.Second*2)
}

func TestMain(m *testing.M) {
	retCode := m.Run()

	os.Exit(retCode)
}
