package dlclient

import (
	"net"
	"os"
	"testing"
	"time"
)

const defaultTestAddress = ":63419"

var (
	defaultTestAddr *net.TCPAddr
	testClient      *Client
)

func init() {
	var err error
	defaultTestAddr, err = net.ResolveTCPAddr("tcp", defaultTestAddress)
	if err != nil {
		panic(err)
	}

	testClient = New(defaultTestAddr, time.Second*3, time.Second*2, time.Second*2)
}

func TestMain(m *testing.M) {
	retCode := m.Run()

	os.Exit(retCode)
}
